package dispatchserver

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"myGinServer/config"

	"golang.org/x/crypto/ssh"
)

// startTestSSHServer 在进程内起一个最小 SSH 服务器，仅用于单元测试：
// 支持 password 与 publickey 两种认证；session 请求统一交给本地 bash -s 执行。
func startTestSSHServer(t *testing.T) (hostPort, user, password string, privKeyPEM []byte) {
	t.Helper()

	_, serverPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	serverSigner, err := ssh.NewSignerFromKey(serverPriv)
	if err != nil {
		t.Fatal(err)
	}

	// 客户端用私钥（用于 publickey 认证测试）
	clientPub, clientPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	clientSSHPub, err := ssh.NewPublicKey(clientPub)
	if err != nil {
		t.Fatal(err)
	}
	clientPEM, err := x509.MarshalPKCS8PrivateKey(clientPriv)
	if err != nil {
		t.Fatal(err)
	}
	privKeyPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: clientPEM})

	const (
		testUser = "tester"
		testPass = "s3cret"
	)
	serverConfig := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == testUser && string(pass) == testPass {
				return nil, nil
			}
			return nil, errors.New("认证失败")
		},
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			if c.User() == testUser && ssh.FingerprintSHA256(pubKey) == ssh.FingerprintSHA256(clientSSHPub) {
				return nil, nil
			}
			return nil, errors.New("未知公钥")
		},
	}
	serverConfig.AddHostKey(serverSigner)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	// 注意 cleanup 按 LIFO 执行：wg.Wait 必须先注册，保证先关监听、后等 accept 协程退出，
	// 否则 Accept 会永远阻塞在未关闭的 listener 上。
	t.Cleanup(wg.Wait)
	t.Cleanup(func() { _ = listener.Close() })

	go func() {
		defer wg.Done()
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go handleTestSSHConn(conn, serverConfig)
		}
	}()

	return listener.Addr().String(), testUser, testPass, privKeyPEM
}

func handleTestSSHConn(conn net.Conn, serverConfig *ssh.ServerConfig) {
	_, chans, reqs, err := ssh.NewServerConn(conn, serverConfig)
	if err != nil {
		return
	}
	go ssh.DiscardRequests(reqs)
	for newChan := range chans {
		if newChan.ChannelType() != "session" {
			_ = newChan.Reject(ssh.UnknownChannelType, "unsupported")
			continue
		}
		ch, chReqs, err := newChan.Accept()
		if err != nil {
			return
		}
		go handleTestSSHSession(ch, chReqs)
	}
}

// handleTestSSHSession 处理 session 的 exec/shell 请求：把命令交给本地 bash -s 执行，
// 标准输出/错误写回 channel，最后回送 exit-status（与 OpenSSH 行为一致）。
func handleTestSSHSession(ch ssh.Channel, chReqs <-chan *ssh.Request) {
	defer ch.Close()
	for req := range chReqs {
		switch req.Type {
		case "exec", "shell":
			var er struct{ Command string }
			if req.Type == "exec" {
				if err := ssh.Unmarshal(req.Payload, &er); err != nil {
					return
				}
			}
			if !req.WantReply {
				continue
			}
			_ = req.Reply(true, nil)

			// 客户端以 exec 形式发送 "bash -s"，脚本经 stdin 传入
			execCmd := exec.Command("bash", "-s")
			execCmd.Stdin = ch
			execCmd.Stdout = ch
			execCmd.Stderr = ch.Stderr() // stderr 走 SSH 扩展数据流，客户端 session.Stderr 才能收到
			exitStatus := uint32(0)
			if err := execCmd.Run(); err != nil {
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) {
					exitStatus = uint32(exitErr.ExitCode())
				} else {
					exitStatus = 1
				}
			}
			_, _ = ch.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{exitStatus}))
			return
		}
	}
}

func newTestRemoteExecutor(t *testing.T, hostPort, user, password string, pemBytes []byte) *remoteExecutor {
	t.Helper()
	cfg := config.SSHConfig{
		Host:           hostPort,
		User:           user,
		Password:       password,
		RsaPrivateKey:  string(pemBytes),
		ShellType:      "bash",
		TimeoutSeconds: 5,
	}
	ex, err := newRemoteExecutor(cfg)
	if err != nil {
		t.Fatalf("newRemoteExecutor 失败: %v", err)
	}
	t.Cleanup(func() { _ = ex.Close() })
	return ex
}

func TestRemoteExecutor_PasswordAuth(t *testing.T) {
	hostPort, user, pass, _ := startTestSSHServer(t)
	ex := newTestRemoteExecutor(t, hostPort, user, pass, nil)

	stdout, stderr, err := ex.RunScript("echo hello-remote; echo second-line")
	if err != nil {
		t.Fatalf("RunScript 失败: %v", err)
	}
	if !strings.Contains(string(stdout), "hello-remote") || !strings.Contains(string(stdout), "second-line") {
		t.Fatalf("stdout 缺少预期内容，实际: %q", string(stdout))
	}
	if len(stderr) != 0 {
		t.Fatalf("stderr 应为空，实际: %q", string(stderr))
	}
}

func TestRemoteExecutor_KeyAuth(t *testing.T) {
	hostPort, user, _, pemBytes := startTestSSHServer(t)
	// 密码留空、只走私钥认证
	ex := newTestRemoteExecutor(t, hostPort, user, "", pemBytes)

	stdout, _, err := ex.RunScript("echo key-ok")
	if err != nil {
		t.Fatalf("RunScript 失败: %v", err)
	}
	if !strings.Contains(string(stdout), "key-ok") {
		t.Fatalf("stdout 缺少预期内容，实际: %q", string(stdout))
	}
}

func TestRemoteExecutor_StderrAndExitCode(t *testing.T) {
	hostPort, user, pass, _ := startTestSSHServer(t)
	ex := newTestRemoteExecutor(t, hostPort, user, pass, nil)

	// 脚本向 stderr 输出并返回非零退出码
	_, stderr, err := ex.RunScript("echo boom >&2; exit 3")
	if err == nil {
		t.Fatal("期望非零退出码返回错误，实际为 nil")
	}
	if !strings.Contains(string(stderr), "boom") {
		t.Fatalf("stderr 应包含 boom，实际: %q", string(stderr))
	}
}

func TestRemoteExecutor_AuthFailure(t *testing.T) {
	hostPort, user, pass, _ := startTestSSHServer(t)
	cfg := config.SSHConfig{
		Host:           hostPort,
		User:           user,
		Password:       "wrong-pass",
		TimeoutSeconds: 3,
	}
	_, err := newRemoteExecutor(cfg)
	if err == nil {
		t.Fatal("期望错误密码连接失败，实际为 nil")
	}
	_ = pass
}

func TestRemoteExecutor_ConfigValidation(t *testing.T) {
	if _, err := newRemoteExecutor(config.SSHConfig{}); err == nil {
		t.Fatal("缺少 host/user 时应返回错误")
	}
	_, err := newRemoteExecutor(config.SSHConfig{Host: "127.0.0.1:1", User: "u", Password: "p", TimeoutSeconds: 1})
	if err == nil {
		t.Fatal("连接不可达时应返回错误")
	}
}
