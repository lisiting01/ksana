package model

import (
	"encoding/json"
	"time"
)

type Duration time.Duration

func (d Duration) String() string {
	return time.Duration(d).String()
}

func (d Duration) MarshalJSON() ([]byte, error) {
	if d == 0 {
		return json.Marshal("")
	}
	return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	if s == "" {
		*d = Duration(0)
		return nil
	}

	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = Duration(dur)
	return nil
}

func (d Duration) ToDuration() time.Duration {
	return time.Duration(d)
}

func DurationFromString(s string) (Duration, error) {
	if s == "" {
		return Duration(0), nil
	}
	dur, err := time.ParseDuration(s)
	return Duration(dur), err
}

func DurationFromTimeDuration(d time.Duration) Duration {
	return Duration(d)
}