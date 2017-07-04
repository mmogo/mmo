package main

import (
	"bytes"
	"image"
	"image/color"

	"github.com/faiface/pixel"
	"github.com/mmogo/mmo/client/assets"
	"github.com/mmogo/mmo/shared"
)

func init() {
	AtlasL = atlasL()
}

// Sprite is an animated sprite
type Sprite struct {
	Picture pixel.Picture
	Frames  map[shared.Direction]map[shared.Action][]pixel.Rect
	Sprite  *pixel.Sprite
	Frame   int     // current frame
	Speed   float64 // frames per second
	elapsed float64
}

func (s *Sprite) Animate(dt float64, facing shared.Direction, action shared.Action) {
	s.elapsed += dt
	if s.Speed == 0 {
		s.Speed = 0.1
	}
	s.Frame = 0 // default frame
	if facing == shared.DIR_NONE {
		facing = DOWN
	}
	if len(s.Frames[facing][action]) > 0 {
		s.Frame = int(s.elapsed/s.Speed) % len(s.Frames[facing][action])
	}
	s.Sprite.Set(s.Picture, s.Frames[facing][action][s.Frame])

}

func (s *Sprite) Draw(target pixel.Target, matrix pixel.Matrix, color color.Color) {
	if s.Sprite == nil {
		s.Sprite = pixel.NewSprite(nil, pixel.Rect{})
	}
	s.Sprite.DrawColorMask(target, matrix, color)
}

// Atlas map
type Atlas map[shared.Direction]map[shared.Action][]pixel.Rect

var AtlasL Atlas

// atlasL for spritesheets generated with Gaurav0's character generator
// https://github.com/Gaurav0/Universal-LPC-Spritesheet-Character-Generator
func atlasL() Atlas {
	m := map[shared.Direction]map[shared.Action][]pixel.Rect{}
	for _, dir := range []shared.Direction{UP, DOWN, LEFT, RIGHT, UPLEFT, UPRIGHT, DOWNLEFT, DOWNRIGHT} {
		m[dir] = make(map[shared.Action][]pixel.Rect)
	}

	bounds := pixel.R(0, 0, 832, 1344)
	frameshigh := 21.00
	frameswide := 13.00
	frameHeight := bounds.Max.Y / frameshigh
	frameWidth := bounds.Max.X / frameswide
	singleframe := pixel.R(0, 0, frameWidth, frameHeight)
	allframes := []pixel.Rect{}
	for y := bounds.Max.Y - frameHeight; y >= 0.00; y = y - frameHeight {
		for x := 0.00; x < bounds.Max.X; x = x + frameWidth {
			frame := singleframe.Moved(pixel.V(x, y))
			allframes = append(allframes, frame)
		}
	}
	// idle
	m[UP][shared.A_IDLE] = []pixel.Rect{allframes[26], allframes[26]}
	m[DOWN][shared.A_IDLE] = []pixel.Rect{allframes[26], allframes[26]}
	m[LEFT][shared.A_IDLE] = []pixel.Rect{allframes[26], allframes[26]}
	m[UPLEFT][shared.A_IDLE] = []pixel.Rect{allframes[26], allframes[26]}
	m[DOWNLEFT][shared.A_IDLE] = []pixel.Rect{allframes[26], allframes[26]}
	m[RIGHT][shared.A_IDLE] = []pixel.Rect{allframes[26], allframes[26]}
	m[UPRIGHT][shared.A_IDLE] = []pixel.Rect{allframes[26], allframes[26]}
	m[DOWNRIGHT][shared.A_IDLE] = []pixel.Rect{allframes[26], allframes[26]}

	// spell casting
	m[UP][shared.A_SPELL] = allframes[0:8]
	m[LEFT][shared.A_SPELL] = allframes[13:21]
	m[UPLEFT][shared.A_SPELL] = allframes[13:21]
	m[DOWNLEFT][shared.A_SPELL] = allframes[13:21]
	m[DOWN][shared.A_SPELL] = allframes[26:33]
	m[RIGHT][shared.A_SPELL] = allframes[39:45]
	m[UPRIGHT][shared.A_SPELL] = allframes[39:45]
	m[DOWNRIGHT][shared.A_SPELL] = allframes[39:45]

	// thrust
	m[UP][shared.A_THRUST] = allframes[52:60]
	m[LEFT][shared.A_THRUST] = allframes[65:73]
	m[UPLEFT][shared.A_THRUST] = allframes[65:73]
	m[DOWNLEFT][shared.A_THRUST] = allframes[65:73]
	m[DOWN][shared.A_THRUST] = allframes[78:86]
	m[RIGHT][shared.A_THRUST] = allframes[91:99]
	m[UPRIGHT][shared.A_THRUST] = allframes[91:99]
	m[DOWNRIGHT][shared.A_THRUST] = allframes[91:99]

	// walk
	m[UP][shared.A_WALK] = allframes[104:113]
	m[LEFT][shared.A_WALK] = allframes[117:126]
	m[UPLEFT][shared.A_WALK] = allframes[117:126]
	m[DOWNLEFT][shared.A_WALK] = allframes[117:126]
	m[DOWN][shared.A_WALK] = allframes[130:139]
	m[RIGHT][shared.A_WALK] = allframes[143:152]
	m[UPRIGHT][shared.A_WALK] = allframes[143:152]
	m[DOWNRIGHT][shared.A_WALK] = allframes[143:152]

	//slash

	m[UP][shared.A_SLASH] = allframes[156:162]
	m[LEFT][shared.A_SLASH] = allframes[169:175]
	m[UPLEFT][shared.A_SLASH] = allframes[169:175]
	m[DOWNLEFT][shared.A_SLASH] = allframes[169:175]
	m[DOWN][shared.A_SLASH] = allframes[182:188]
	m[RIGHT][shared.A_SLASH] = allframes[195:201]
	m[UPRIGHT][shared.A_SLASH] = allframes[195:201]
	m[DOWNRIGHT][shared.A_SLASH] = allframes[195:201]

	//shoot
	m[UP][shared.A_SHOOT] = allframes[208:221]
	m[LEFT][shared.A_SHOOT] = allframes[221:234]
	m[UPLEFT][shared.A_SHOOT] = allframes[221:234]
	m[DOWNLEFT][shared.A_SHOOT] = allframes[221:234]
	m[DOWN][shared.A_SHOOT] = allframes[234:247]
	m[RIGHT][shared.A_SHOOT] = allframes[247:260]
	m[UPRIGHT][shared.A_SHOOT] = allframes[247:260]
	m[DOWNRIGHT][shared.A_SHOOT] = allframes[247:260]

	//hurt
	m[UP][shared.A_HURT] = allframes[260:265]
	m[LEFT][shared.A_HURT] = allframes[260:265]
	m[UPLEFT][shared.A_HURT] = allframes[260:265]
	m[DOWNLEFT][shared.A_HURT] = allframes[260:265]
	m[DOWN][shared.A_HURT] = allframes[260:265]
	m[RIGHT][shared.A_HURT] = allframes[260:265]

	//dead
	m[UP][shared.A_DEAD] = allframes[265:]
	m[LEFT][shared.A_DEAD] = allframes[265:]
	m[UPLEFT][shared.A_DEAD] = allframes[265:]
	m[DOWNLEFT][shared.A_DEAD] = allframes[265:]
	m[DOWN][shared.A_DEAD] = allframes[265:]
	m[RIGHT][shared.A_DEAD] = allframes[265:]
	m[UPRIGHT][shared.A_DEAD] = allframes[265:]
	m[DOWNRIGHT][shared.A_DEAD] = allframes[265:]
	return m
}

// LoadSpriteSheet returns an animated sprite derived from image path and atlas function
func LoadSpriteSheet(path string, atlas Atlas) (*Sprite, error) {
	pic, err := loadPicture(path)
	if err != nil {
		return nil, err
	}
	s := new(Sprite)
	s.Frames = AtlasL
	if atlas != nil {
		s.Frames = atlas
	}
	s.Sprite = pixel.NewSprite(nil, pixel.Rect{})
	s.Picture = pic
	return s, nil
}

// loadPicture from assets
func loadPicture(path string) (pixel.Picture, error) {
	contents, err := assets.Asset(path)
	if err != nil {
		return nil, err
	}
	file := bytes.NewReader(contents)
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}
