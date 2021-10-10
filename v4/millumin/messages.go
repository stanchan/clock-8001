package millumin

import (
	"github.com/stanchan/go-osc/osc"
)

// MediaInfo contains the information of a given media file in Millumin playback state
type MediaInfo struct {
	Index    int32
	Name     string
	Duration float32
}

// UnmarshalOSC converts a osc.Message to MediaInfo message
func (i *MediaInfo) UnmarshalOSC(msg *osc.Message) error {
	if err := msg.UnmarshalArgument(0, &i.Index); err != nil {
		return err
	}
	if err := msg.UnmarshalArgument(1, &i.Name); err != nil {
		return err
	}
	if err := msg.UnmarshalArgument(2, &i.Duration); err != nil {
		return err
	}

	return nil
}

// MediaTime constains the Millumin media playback timestamps from
// /millumin/layer:time/media/time f:value, f:duration osc messages
type MediaTime struct {
	Value    float32
	Duration float32
}

// UnmarshalOSC converts a osc.Message to MediaTime
func (mt *MediaTime) UnmarshalOSC(msg *osc.Message) error {
	if err := msg.UnmarshalArgument(0, &mt.Value); err != nil {
		return err
	}
	if err := msg.UnmarshalArgument(1, &mt.Duration); err != nil {
		return err
	}

	return nil
}

// MediaStarted constains information of new media playback start
// from /millumin/layer:time/mediaStarted i:index s:name f:duration
type MediaStarted struct {
	MediaInfo
}

// UnmarshalOSC converts osc.Message to MediaStarted
func (ms *MediaStarted) UnmarshalOSC(msg *osc.Message) error {
	return ms.MediaInfo.UnmarshalOSC(msg)
}

// MediaPaused is a pause event from Millumin in message
// /millumin/layer:time/mediaPaused i:index s:name f:duration
type MediaPaused struct {
	MediaInfo
}

// UnmarshalOSC converts osc.Message to MediaPaused
func (ms *MediaPaused) UnmarshalOSC(msg *osc.Message) error {
	return ms.MediaInfo.UnmarshalOSC(msg)
}

// MediaStopped is a playback stop message from Millumin
// /millumin/layer:time/mediaStopped i:index s:name f:duration
type MediaStopped struct {
	MediaInfo
}

// UnmarshalOSC converts osc.Message to MediaStopped
func (ms *MediaStopped) UnmarshalOSC(msg *osc.Message) error {
	return ms.MediaInfo.UnmarshalOSC(msg)
}
