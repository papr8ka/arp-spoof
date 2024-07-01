package interactive

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/papr8ka/arp-spoof/arp"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
	"image"
	"math"
	"net"
)

type input int

const (
	inputTargetMAC input = iota
	inputSpoofedIP
	inputSpoofedMAC

	inputs
)

var (
	title = [inputs]string{
		"Target MAC",
		"Spoofed IP",
		"Spoofed MAC",
	}
)

const (
	rectangleHeight        = 50
	rectangleVerticalSpace = 30

	textZoom = 4
)

type element struct {
	text      string
	rectangle image.Rectangle
	overed    bool
}

type interactive struct {
	arp arp.Arp

	font text.Face

	element [inputs]element

	editedInput input
	editedText  string
}

func New(arp arp.Arp) ebiten.Game {
	return &interactive{
		arp: arp,

		font: text.NewGoXFace(basicfont.Face7x13),

		editedInput: inputs,
	}
}

func (i *interactive) getInputValue(input input) string {
	if i.editedInput == input {
		return i.editedText
	}

	switch input {
	case inputSpoofedIP:
		return i.arp.GetSpoofedIP()
	case inputSpoofedMAC:
		return i.arp.GetSpoofedMAC()
	case inputTargetMAC:
		return i.arp.GetTargetMAC()

	default:
		return ""
	}
}

func (i *interactive) setInputValue(input input, newValue string) {
	spoofedIP := i.arp.GetSpoofedIP()
	spoofedMAC := i.arp.GetSpoofedMAC()
	targetMAC := i.arp.GetTargetMAC()

	switch input {
	case inputSpoofedIP:
		spoofedIP = newValue
	case inputSpoofedMAC:
		spoofedMAC = newValue
	case inputTargetMAC:
		targetMAC = newValue

	default:
		return
	}

	_ = i.arp.SetParameter(targetMAC, spoofedIP, spoofedMAC)
}

const (
	MACMaximum = int64(0xFFF7FFFFFFFF)
)

func joinByteArrayIntoData(v []byte, count int) (result int64) {
	for index := 0; index < count; index++ {
		result |= int64(v[index]) << ((count - index - 1) * 8)
	}
	return
}

func cutDataIntoByteArray(v int64, count int) []byte {
	result := make([]byte, count)
	for index := 0; index < count; index++ {
		result[index] = byte((v >> ((count - index - 1) * 8)) & 0xFF)
	}
	return result
}

func (i *interactive) increaseDecreaseIP(value string, delta int) string {
	if ip := net.ParseIP(value); ip == nil {
		return value
	} else {
		return net.IP(cutDataIntoByteArray(joinByteArrayIntoData(ip.To4(), 4)+int64(delta), 4)).String()
	}
}

func (i *interactive) increaseDecreaseMAC(value string, delta int) string {
	if mac, err := net.ParseMAC(value); err == nil {
		v := (joinByteArrayIntoData(mac, 6) + int64(delta)) % MACMaximum
		if v < 0 {
			v = MACMaximum
		}
		return net.HardwareAddr(cutDataIntoByteArray(v, 6)).String()
	} else {
		return value
	}
}

func (i *interactive) increaseDecreaseInput(input input, delta int) {
	switch input {
	case inputTargetMAC:
		fallthrough
	case inputSpoofedMAC:
		i.setInputValue(input, i.increaseDecreaseMAC(i.getInputValue(input), delta))

	case inputSpoofedIP:
		i.setInputValue(input, i.increaseDecreaseIP(i.getInputValue(input), delta))

	default:
	}
}

func (i *interactive) Update() error {
	windowWidth, windowHeight := ebiten.WindowSize()
	cursorX, cursorY := ebiten.CursorPosition()

	inputHeight := int(rectangleHeight*inputs + rectangleVerticalSpace*(inputs-1))

	verticalBase := windowHeight/2 - inputHeight/2
	for currentInput := input(0); currentInput < inputs; currentInput++ {
		e := &i.element[currentInput]
		e.text = fmt.Sprintf("%s: %s", title[currentInput], i.getInputValue(currentInput))
		width, _ := text.Measure(e.text, i.font, 0)
		width *= textZoom
		e.rectangle = image.Rect(windowWidth/2-int(width)/2, verticalBase, windowWidth/2-int(width)/2+int(width), verticalBase+rectangleHeight)
		verticalBase += rectangleVerticalSpace + rectangleHeight

		e.overed = e.rectangle.Overlaps(image.Rect(cursorX, cursorY, cursorX+1, cursorY+1))
	}

	for currentInput := input(0); currentInput < inputs; currentInput++ {
		if e := &i.element[currentInput]; e.overed && i.editedInput == inputs {
			if _, wheel := ebiten.Wheel(); wheel != 0 {
				i.increaseDecreaseInput(currentInput, int(math.Ceil(wheel)))
			}

			if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
				i.editedText = i.getInputValue(currentInput)
				i.editedInput = currentInput
			}
		}
	}

	if i.editedInput != inputs {
		if ebiten.IsKeyPressed(ebiten.KeyEscape) {
			i.editedText = ""
			i.editedInput = inputs
		}

		if ebiten.IsKeyPressed(ebiten.KeyDelete) {
			i.editedText = ""
		}

		if ebiten.IsKeyPressed(ebiten.KeyEnter) || ebiten.IsKeyPressed(ebiten.KeyKPEnter) {
			i.setInputValue(i.editedInput, i.editedText)
			i.editedInput = inputs
			i.editedText = ""
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
			if len(i.editedText) > 1 {
				i.editedText = i.editedText[:len(i.editedText)-1]
			} else {
				i.editedText = ""
			}
		}

		i.editedText = string(ebiten.AppendInputChars([]rune(i.editedText)))
	}

	return nil
}

func (i *interactive) Draw(screen *ebiten.Image) {
	for currentInput := input(0); currentInput < inputs; currentInput++ {
		e := &i.element[currentInput]
		transformation := ebiten.DrawImageOptions{}
		transformation.GeoM.Scale(textZoom, textZoom)
		transformation.GeoM.Translate(
			float64(e.rectangle.Bounds().Min.X+e.rectangle.Dx()/2),
			float64(e.rectangle.Bounds().Min.Y),
		)
		if i.editedInput == inputs {
			if e.overed {
				transformation.ColorScale.ScaleWithColor(colornames.Red)
			}
		} else {
			if i.editedInput == currentInput {
				transformation.ColorScale.ScaleWithColor(colornames.Green)
			}
		}
		text.Draw(screen, e.text, i.font, &text.DrawOptions{
			DrawImageOptions: transformation,
			LayoutOptions: text.LayoutOptions{
				PrimaryAlign: text.AlignCenter,
			},
		})
	}
}

func (i *interactive) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}
