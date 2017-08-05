package millumin

import (
	"github.com/hypebeast/go-osc/osc"
)

type MediaInfo struct {
	Index    int32
	Name     string
	Duration float32
}

func (i *MediaInfo) UnmarshalOSC(msg *osc.Message) error {
	if err := unmarshalArgument(msg, 0, &i.Index); err != nil {
		return err
	}
	if err := unmarshalArgument(msg, 1, &i.Name); err != nil {
		return err
	}
	if err := unmarshalArgument(msg, 2, &i.Duration); err != nil {
		return err
	}

	return nil
}

// /millumin/layer:time/media/time f:value, f:duration
type MediaTime struct {
	Value    float32
	Duration float32
}

func (mt *MediaTime) UnmarshalOSC(msg *osc.Message) error {
	if err := unmarshalArgument(msg, 0, &mt.Value); err != nil {
		return err
	}
	if err := unmarshalArgument(msg, 1, &mt.Duration); err != nil {
		return err
	}

	return nil
}

// /millumin/layer:time/mediaStarted i:index s:name f:duration
type MediaStarted struct {
	MediaInfo
}

func (ms *MediaStarted) UnmarshalOSC(msg *osc.Message) error {
	return ms.MediaInfo.UnmarshalOSC(msg)
}

// /millumin/layer:time/mediaPaused i:index s:name f:duration
type MediaPaused struct {
	MediaInfo
}

func (ms *MediaPaused) UnmarshalOSC(msg *osc.Message) error {
	return ms.MediaInfo.UnmarshalOSC(msg)
}

// /millumin/layer:time/mediaStopped i:index s:name f:duration
type MediaStopped struct {
	MediaInfo
}

func (ms *MediaStopped) UnmarshalOSC(msg *osc.Message) error {
	return ms.MediaInfo.UnmarshalOSC(msg)
}
