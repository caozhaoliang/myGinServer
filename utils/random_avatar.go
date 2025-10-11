package utils

import (
	"fmt"
	"math/rand"
)

// 生成随机肤色
func randomSkinColor() string {
	colors := []string{
		"#FFDBAC", "#F1C27D", "#E0AC69", "#C68642", "#8D5B2F",
		"#D9B38C", "#A67B5B", "#734A12", "#5C4033", "#3E2723",
	}
	return colors[rand.Intn(len(colors))]
}

// 生成随机发色
func randomHairColor() string {
	colors := []string{
		"#000000", "#333333", "#555555", "#777777",
		"#FFD700", "#8B4513", "#A52A2A", "#FF6347",
	}
	return colors[rand.Intn(len(colors))]
}

// 生成随机眼睛颜色
func randomEyeColor() string {
	colors := []string{
		"#2E8B57", "#4682B4", "#8B4513", "#A0522D",
		"#800080", "#FF6347", "#FFD700", "#00CED1",
	}
	return colors[rand.Intn(len(colors))]
}

// 生成随机发型
func randomHairStyle(x, y, width, height int) string {
	style := rand.Intn(3)
	switch style {
	case 0: // 短发
		return fmt.Sprintf(`<path d="M%d,%d Q%d,%d %d,%d L%d,%d Q%d,%d %d,%d Z" fill="%s"/>`,
			x-10, y+20, x, y-30, x+width+10, y+20,
			x+width+10, y+height/3, x, y-10, x-10, y+height/3,
			randomHairColor())
	case 1: // 长发
		return fmt.Sprintf(`<path d="M%d,%d Q%d,%d %d,%d L%d,%d L%d,%d Q%d,%d %d,%d Z" fill="%s"/>`,
			x-20, y+20, x, y-40, x+width+20, y+20,
			x+width+20, y+height+30, x-20, y+height+30,
			x, y-20, x-20, y+20,
			randomHairColor())
	case 2: // 卷发
		return fmt.Sprintf(`
			<path d="M%d,%d Q%d,%d %d,%d L%d,%d Q%d,%d %d,%d Z" fill="%s"/>
			<circle cx="%d" cy="%d" r="%d" fill="%s"/>
			<circle cx="%d" cy="%d" r="%d" fill="%s"/>
			<circle cx="%d" cy="%d" r="%d" fill="%s"/>
		`,
			x-15, y+15, x, y-35, x+width+15, y+15,
			x+width+15, y+height/2, x, y-15, x-15, y+height/2,
			randomHairColor(),
			x-5, y+10, 10, randomHairColor(),
			x+width+5, y+10, 10, randomHairColor(),
			x+width/2, y-10, 12, randomHairColor())
	}
	return ""
}

// 生成随机眼睛
func randomEyes(x, y, width int) string {
	eyeSize := rand.Intn(5) + 6
	eyeStyle := rand.Intn(2)

	leftX := x + width/4
	rightX := x + width*3/4

	if eyeStyle == 0 { // 圆形眼睛
		return fmt.Sprintf(`
			<circle cx="%d" cy="%d" r="%d" fill="white"/>
			<circle cx="%d" cy="%d" r="%d" fill="%s"/>
			<circle cx="%d" cy="%d" r="%d" fill="black"/>
			
			<circle cx="%d" cy="%d" r="%d" fill="white"/>
			<circle cx="%d" cy="%d" r="%d" fill="%s"/>
			<circle cx="%d" cy="%d" r="%d" fill="black"/>
		`,
			leftX, y, eyeSize,
			leftX, y, eyeSize/2, randomEyeColor(),
			leftX-eyeSize/4, y-eyeSize/4, eyeSize/4,

			rightX, y, eyeSize,
			rightX, y, eyeSize/2, randomEyeColor(),
			rightX-eyeSize/4, y-eyeSize/4, eyeSize/4)
	} else { // 椭圆形眼睛
		return fmt.Sprintf(`
			<ellipse cx="%d" cy="%d" rx="%d" ry="%d" fill="white"/>
			<ellipse cx="%d" cy="%d" rx="%d" ry="%d" fill="%s"/>
			<ellipse cx="%d" cy="%d" rx="%d" ry="%d" fill="black"/>
			
			<ellipse cx="%d" cy="%d" rx="%d" ry="%d" fill="white"/>
			<ellipse cx="%d" cy="%d" rx="%d" ry="%d" fill="%s"/>
			<ellipse cx="%d" cy="%d" rx="%d" ry="%d" fill="black"/>
		`,
			leftX, y, eyeSize, eyeSize*2/3,
			leftX, y, eyeSize/2, eyeSize/2, randomEyeColor(),
			leftX-eyeSize/4, y-eyeSize/4, eyeSize/4, eyeSize/4,

			rightX, y, eyeSize, eyeSize*2/3,
			rightX, y, eyeSize/2, eyeSize/2, randomEyeColor(),
			rightX-eyeSize/4, y-eyeSize/4, eyeSize/4, eyeSize/4)
	}
}

// 生成随机嘴巴
func randomMouth(x, y, width int) string {
	mouthStyle := rand.Intn(3)
	mouthWidth := rand.Intn(20) + 30

	switch mouthStyle {
	case 0: // 微笑
		return fmt.Sprintf(`<path d="M%d,%d Q%d,%d %d,%d" stroke="%s" stroke-width="%d" fill="none"/>`,
			x+(width-mouthWidth)/2, y,
			x+width/2, y+10,
			x+width-(width-mouthWidth)/2, y,
			"#8B4513", rand.Intn(3)+2)
	case 1: // 中性
		return fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="%d"/>`,
			x+(width-mouthWidth)/2, y,
			x+width-(width-mouthWidth)/2, y,
			"#8B4513", rand.Intn(3)+2)
	case 2: // 皱眉
		return fmt.Sprintf(`<path d="M%d,%d Q%d,%d %d,%d" stroke="%s" stroke-width="%d" fill="none"/>`,
			x+(width-mouthWidth)/2, y,
			x+width/2, y-10,
			x+width-(width-mouthWidth)/2, y,
			"#8B4513", rand.Intn(3)+2)
	}
	return ""
}

// 生成随机鼻子
func randomNose(x, y, width int) string {
	noseStyle := rand.Intn(2)
	noseSize := rand.Intn(5) + 5

	switch noseStyle {
	case 0: // 三角形鼻子
		return fmt.Sprintf(`<polygon points="%d,%d %d,%d %d,%d" fill="%s" opacity="0.8"/>`,
			x+width/2, y,
			x+width/2-noseSize, y+noseSize*2,
			x+width/2+noseSize, y+noseSize*2,
			"#D2B48C")
	case 1: // 圆形鼻子
		return fmt.Sprintf(`<circle cx="%d" cy="%d" r="%d" fill="%s" opacity="0.8"/>`,
			x+width/2, y+noseSize,
			noseSize,
			"#D2B48C")
	}
	return ""
}

// 生成随机人脸头像
func generateHumanAvatar(width, height int) string {
	// SVG头部
	svg := fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, width, height)

	// 背景
	svg += `<rect width="100%" height="100%" fill="#f0f0f0"/>`

	// 脸部位置和尺寸
	faceX := width / 4
	faceY := height / 5
	faceWidth := width / 2
	faceHeight := height * 3 / 5

	// 发型
	svg += randomHairStyle(faceX, faceY, faceWidth, faceHeight)

	// 脸部轮廓
	faceRadius := faceWidth / 2
	svg += fmt.Sprintf(`<ellipse cx="%d" cy="%d" rx="%d" ry="%d" fill="%s"/>`,
		width/2, faceY+faceHeight/2,
		faceRadius, faceHeight/2,
		randomSkinColor())

	// 眼睛
	eyeY := faceY + faceHeight/3
	svg += randomEyes(faceX, eyeY, faceWidth)

	// 鼻子
	noseY := faceY + faceHeight/2 - 10
	svg += randomNose(faceX, noseY, faceWidth)

	// 嘴巴
	mouthY := faceY + faceHeight*3/4 - 10
	svg += randomMouth(faceX, mouthY, faceWidth)

	// 随机添加眼镜
	if rand.Intn(3) == 0 {
		glassSize := faceWidth/2 + 10
		svg += fmt.Sprintf(`
			<circle cx="%d" cy="%d" r="%d" fill="none" stroke="#333" stroke-width="%d"/>
			<circle cx="%d" cy="%d" r="%d" fill="none" stroke="#333" stroke-width="%d"/>
			<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#333" stroke-width="%d"/>
			<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#333" stroke-width="%d"/>
			<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#333" stroke-width="%d"/>
		`,
			faceX+faceWidth/4, eyeY, glassSize/4, 2,
			faceX+faceWidth*3/4, eyeY, glassSize/4, 2,
			faceX+faceWidth/4+glassSize/4, eyeY, faceX+faceWidth*3/4-glassSize/4, eyeY, 2,
			faceX+faceWidth/4-glassSize/4, eyeY, faceX+faceWidth/4-glassSize/8, eyeY-glassSize/8, 2,
			faceX+faceWidth*3/4+glassSize/4, eyeY, faceX+faceWidth*3/4+glassSize/8, eyeY-glassSize/8, 2)
	}

	// SVG尾部
	svg += `</svg>`

	// 写入文件
	return svg
}
