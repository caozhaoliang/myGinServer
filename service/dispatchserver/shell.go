package dispatchserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"myGinServer/api/response"
	"myGinServer/config"
	mdispatch "myGinServer/models/dispatch"

	"golang.org/x/crypto/ssh"
)

// remoteExecutor 远程执行器：参照 bigdata-faas 的 processor.go 与 gopkg/ssh 实现。
// 通过 SSH 连接远程主机执行 shell 文本，支持密码 / 私钥（可带口令）两种认证；
// 执行方式为把脚本经 stdin 交给远程解释器（sh/bash），捕获 stdout 与 stderr。
type remoteExecutor struct {
	cfg    config.SSHConfig
	client *ssh.Client
}

// newRemoteExecutor 建立 SSH 连接。认证方式与 gopkg/ssh.NewSSHClient 一致：
// 密码 + 可选私钥；私钥支持文件路径或内联 PEM。
func newRemoteExecutor(cfg config.SSHConfig) (*remoteExecutor, error) {
	if cfg.Host == "" || cfg.User == "" {
		return nil, fmt.Errorf("ssh 配置缺失：host / user 必填")
	}

	auth := make([]ssh.AuthMethod, 0, 2)
	if cfg.Password != "" {
		auth = append(auth, ssh.Password(cfg.Password))
	}
	if cfg.RsaPrivateKey != "" || cfg.RsaPrivateKeyPath != "" {
		keyPEM := []byte(cfg.RsaPrivateKey)
		if len(keyPEM) == 0 {
			data, err := os.ReadFile(cfg.RsaPrivateKeyPath)
			if err != nil {
				return nil, fmt.Errorf("读取 ssh 私钥文件失败: %w", err)
			}
			keyPEM = data
		}
		var (
			signer ssh.Signer
			err    error
		)
		if cfg.Passphrase == "" {
			signer, err = ssh.ParsePrivateKey(keyPEM)
		} else {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(keyPEM, []byte(cfg.Passphrase))
		}
		if err != nil {
			return nil, fmt.Errorf("解析 ssh 私钥失败: %w", err)
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}
	if len(auth) == 0 {
		return nil, fmt.Errorf("ssh 配置缺失：password 与私钥至少提供一种认证方式")
	}

	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	client, err := ssh.Dial("tcp", cfg.Host, &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            auth,
		Timeout:         timeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 与 gopkg/ssh 一致；生产环境建议改为 known_hosts 校验
	})
	if err != nil {
		return nil, fmt.Errorf("ssh 连接 %s@%s 失败: %w", cfg.User, cfg.Host, err)
	}
	return &remoteExecutor{cfg: cfg, client: client}, nil
}

func (e *remoteExecutor) Close() error {
	return e.client.Close()
}

// RunScript 将脚本经 stdin 交给远程解释器执行（等价于 ssh host 'bash -s' < script）。
// 返回标准输出与标准错误；脚本退出码非 0 时 err 为 *ssh.ExitError。
func (e *remoteExecutor) RunScript(script string) (stdout, stderr []byte, err error) {
	session, err := e.client.NewSession()
	if err != nil {
		return nil, nil, fmt.Errorf("创建 ssh session 失败: %w", err)
	}
	defer session.Close()

	shellType := e.cfg.ShellType
	if shellType == "" {
		shellType = "bash"
	}
	var out, errBuf bytes.Buffer
	session.Stdout = &out
	session.Stderr = &errBuf
	session.Stdin = strings.NewReader(script)
	if err := session.Run(shellType + " -s"); err != nil {
		return out.Bytes(), errBuf.Bytes(), fmt.Errorf("远程命令执行失败: %w", err)
	}
	return out.Bytes(), errBuf.Bytes(), nil
}

// runShell 消费 exec_queue 中的 shell 运行项：SSH 连接远程主机执行脚本并回写结果。
// 与 runSQL 走同一套队列状态机（pending/running/success/failed），结果写入 response.output。
func (n *NodeServer) runShell(ctx context.Context, entity TestRunEntity) {
	queue, errQuery := n.store.QueryExecQueue(ctx, entity.TenantDb, entity.RunId)
	if errQuery != nil {
		return
	}
	mu.Lock()
	if mdispatch.EntityAlreadyRun(queue.Status) {
		mu.Unlock()
		return
	}
	_ = n.store.UpdateExecQueueResp(ctx, entity.TenantDb, entity.RunId, "{}", "running")
	mu.Unlock()

	resp := response.TestRunResp{Sql: entity.Sql, Msg: "success"}
	status := "success"

	executor, err := newRemoteExecutor(config.GetConfig().SSH)
	if err != nil {
		resp.Msg = err.Error()
		status = "failed"
	} else {
		// 执行完立即关闭连接（不复用长连接，避免连接数膨胀）
		defer executor.Close()
		stdout, stderr, runErr := executor.RunScript(entity.Sql)
		resp.Output = strings.TrimSpace(string(stdout) + "\n" + string(stderr))
		if runErr != nil {
			resp.Msg = runErr.Error()
			status = "failed"
		}
	}

	payload, _ := json.Marshal(resp)
	if err := n.store.UpdateExecQueueResp(ctx, entity.TenantDb, entity.RunId, string(payload), status); err != nil {
		fmt.Println(err.Error())
	}
}
