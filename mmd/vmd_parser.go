package mmd

import (
	"bytes"
	"fmt"
	"io"
	"sort"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// VMDParser is parser for .vmd animation.
type VMDParser struct {
	baseParser
	header *Header
}

type Animation struct {
	Name   string
	Bone   []*AnimationBoneSample
	Morph  []*AnimationMorphSample
	Camera []*AnimationCameraSample
	Light  []*AnimationLightSample
}

type AnimationBoneSample struct {
	Target   string
	Frame    int
	Position Vector3
	Rotation Quaternion
	Params   [64]byte
}

type AnimationMorphSample struct {
	Target string
	Frame  int
	Value  float32
}

type AnimationCameraSample struct {
	Frame      int
	Distance   float32
	Position   Vector3
	Rotation   Vector3 // 4?
	Params     [24]byte
	FoV        float32
	Projection byte
}

type AnimationLightSample struct {
	Frame    int
	Color    Vector3
	Position Vector3
}

type BoneChannel struct {
	Target    string
	Keyframes []*AnimationBoneSample
	Frames    []uint32
	Rotations []*Quaternion
	Positions []*Vector3
}

type MorphChannel struct {
	Target    string
	Keyframes []*AnimationMorphSample
	Frames    []uint32
	Weights   []float32
}

func (a *Animation) GetBoneChannels() map[string]*BoneChannel {
	sort.Slice(a.Bone, func(i, j int) bool { return a.Bone[i].Frame < a.Bone[j].Frame })

	r := map[string]*BoneChannel{}
	for _, s := range a.Bone {
		a, ok := r[s.Target]
		if !ok {
			a = &BoneChannel{Target: s.Target}
			r[s.Target] = a
		}
		a.Frames = append(a.Frames, uint32(s.Frame))
		a.Rotations = append(a.Rotations, &s.Rotation)
		a.Positions = append(a.Positions, &s.Position)
	}
	return r
}

func (a *Animation) GetMorphChannels() map[string]*MorphChannel {
	sort.Slice(a.Morph, func(i, j int) bool { return a.Morph[i].Frame < a.Morph[j].Frame })

	r := map[string]*MorphChannel{}
	for _, s := range a.Morph {
		a, ok := r[s.Target]
		if !ok {
			a = &MorphChannel{Target: s.Target}
			r[s.Target] = a
		}
		a.Frames = append(a.Frames, uint32(s.Frame))
		a.Weights = append(a.Weights, s.Value)
	}
	return r
}

// NewVMDParser returns new parser.
func NewVMDParser(r io.Reader) *VMDParser {
	return &VMDParser{baseParser: baseParser{r: r}}
}

// Parse animation data.
func (p *VMDParser) Parse() (*Animation, error) {
	var anim Animation
	var supportedFormat = "Vocaloid Motion Data 0002"

	formatName := p.readString(30)
	if formatName != supportedFormat {
		return nil, fmt.Errorf("Format error: %v != %v", formatName, supportedFormat)
	}

	anim.Name = p.readString(20)

	frames := p.readInt()
	for i := 0; i < frames; i++ {
		sample := &AnimationBoneSample{}
		sample.Target = p.readString(15)
		sample.Frame = p.readInt()
		p.read(&sample.Position)
		p.read(&sample.Rotation)
		p.read(&sample.Params)
		anim.Bone = append(anim.Bone, sample)
	}

	frames = p.readInt()
	for i := 0; i < frames; i++ {
		sample := &AnimationMorphSample{}
		sample.Target = p.readString(15)
		sample.Frame = p.readInt()
		p.read(&sample.Value)
		anim.Morph = append(anim.Morph, sample)
	}

	return &anim, p.err
}

func (p *VMDParser) readString(len int) string {
	b := make([]byte, len)
	_ = p.read(b)
	utf8Data, _, _ := transform.Bytes(japanese.ShiftJIS.NewDecoder(), bytes.SplitN(b, []byte{0}, 2)[0])
	return string(utf8Data)
}
