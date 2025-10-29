package captcha

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/gorilla/sessions" // 用于Session管理
)

var (
	store         = sessions.NewCookieStore([]byte("my-app-access-key")) // Session存储
	chars         = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	font          *truetype.Font
	captchaWidth  = 200 // 验证码图片宽度
	captchaHeight = 80  // 验证码图片高度
)

func init() {
	// 加载字体
	fontBytes, err := os.ReadFile("~/Downloads/包图小白体.ttf")
	if err != nil {
		panic(err)
	}
	font, err = freetype.ParseFont(fontBytes)
	if err != nil {
		panic(err)
	}
}

// 生成随机字符串
func randomStr(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

// 生成验证码图片
func generateCaptcha(w http.ResponseWriter, r *http.Request) {
	width, height := 200, 80
	// 创建RGBA图像（透明背景）
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// 填充白色背景
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// 生成随机字符
	captchaText := randomStr(4)
	// 存储到Session
	session, _ := store.Get(r, "captcha-session")
	session.Values["code"] = captchaText
	session.Options.MaxAge = 60 * 5 // 5分钟过期
	session.Save(r, w)

	// 绘制字符（简化版，实际需字体库支持，如使用 github.com/golang/freetype）
	// 3. 初始化freetype上下文
	ctx := freetype.NewContext()
	ctx.SetDPI(72)      // 标准DPI
	ctx.SetFont(font)   // 加载的字体
	ctx.SetFontSize(36) // 字体大小
	ctx.SetClip(img.Bounds())
	ctx.SetDst(img)
	ctx.SetSrc(&image.Uniform{color.Black}) // 默认颜色（后续会覆盖）

	// 这里用简单的矩形模拟字符（实际需加载字体绘制）
	for i, ch := range captchaText {
		// 随机颜色（深色）
		c := color.RGBA{
			R: uint8(rand.Intn(80) + 30),
			G: uint8(rand.Intn(80) + 30),
			B: uint8(rand.Intn(80) + 30),
		}

		ctx.SetSrc(&image.Uniform{c})
		// 随机位置偏移（避免字符重叠，增加识别难度）
		x := 10 + i*45 + rand.Intn(10)
		y := captchaHeight/2 + rand.Intn(15)
		// 绘制字符
		_, err := ctx.DrawString(string(ch), freetype.Pt(x, y))
		if err != nil {
			http.Error(w, "字符绘制失败", http.StatusInternalServerError)
			return
		}
	}

	// 添加干扰线
	for i := 0; i < 6; i++ {
		c := color.RGBA{
			R: uint8(rand.Intn(100) + 50),
			G: uint8(rand.Intn(100) + 50),
			B: uint8(rand.Intn(100) + 50),
		}
		x1, y1 := rand.Intn(width), rand.Intn(height)
		x2, y2 := rand.Intn(width), rand.Intn(height)
		// 绘制线段（简化版，实际需用绘图库）
		for t := 0; t < 100; t++ {
			f := float64(t) / 100
			x := int(float64(x1)*(1-f) + float64(x2)*f)
			y := int(float64(y1)*(1-f) + float64(y2)*f)
			img.Set(x, y, c)
		}
	}

	// 输出图片
	w.Header().Set("Content-Type", "image/png")
	png.Encode(w, img)
}

// 验证接口
func verifyCaptcha(w http.ResponseWriter, r *http.Request) {
	input := r.URL.Query().Get("code")
	session, _ := store.Get(r, "captcha-session")
	code, ok := session.Values["code"].(string)
	if !ok || input != code {
		w.Write([]byte("验证失败"))
		return
	}
	w.Write([]byte("验证通过"))
}
