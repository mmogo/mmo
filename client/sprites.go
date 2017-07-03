package main

import (
	"bytes"
	"image"
	"log"

	"github.com/faiface/pixel"
	"github.com/mmogo/mmo/client/assets"
	"github.com/mmogo/mmo/shared"
)

// Sprite is an animated sprite
type Sprite struct {
	Picture pixel.Picture
	Frames  map[shared.Direction]map[shared.Action][]pixel.Rect
	Sprite  *pixel.Sprite
	Frame   int // current frame
}

type Atlas func() map[shared.Direction]map[shared.Action][]pixel.Rect

var AtlasDefault = atlasDefault

func atlasDefault() map[shared.Direction]map[shared.Action][]pixel.Rect {
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
	log.Println(singleframe)
	allframes := []pixel.Rect{}
	for y := bounds.Max.Y - frameHeight; y >= 0.00; y = y - frameHeight {
		for x := 0.00; x < bounds.Max.X; x = x + frameWidth {
			frame := singleframe.Moved(pixel.V(x, y))
			allframes = append(allframes, frame)
		}
	}
	log.Println("total:", len(allframes))

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

	for i, f := range allframes {
		log.Println(i, f)
	}
	return m
}
func LoadSpriteSheet(path string, atlas Atlas) (*Sprite, error) {
	pic, err := loadPicture(path)
	if err != nil {
		return nil, err
	}
	if atlas == nil {
		atlas = AtlasDefault
	}
	s := new(Sprite)
	s.Sprite = pixel.NewSprite(nil, pixel.Rect{})
	s.Frames = atlas()
	s.Picture = pic
	return s, nil
}

func loadPicture(path string) (pixel.Picture, error) {
	contents, err := assets.Asset(path)
	if err != nil {
		return nil, err
	}
	file := bytes.NewBuffer(contents)
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}
