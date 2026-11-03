package online

import "strconv"

// Online game modes
const (
	ModeWordDrop     = "word_drop"
	ModeSentenceRace = "sentence_race"
	ModeWordPool     = "word_pool"
)

// Modes lists the modes a host can choose from, in menu order
var Modes = []string{ModeWordDrop, ModeSentenceRace, ModeWordPool}

// ModeName returns the display name of a mode
func ModeName(mode string) string {
	switch mode {
	case ModeWordDrop:
		return "Word Drop"
	case ModeSentenceRace:
		return "Sentence Race"
	case ModeWordPool:
		return "Word Pool"
	}
	return mode
}

// Setting keys carried by set_option
const (
	OptionMode       = "mode"
	OptionCount      = "count"
	OptionTimeLimit  = "time_limit"
	OptionSpawnRate  = "spawn_rate"
	OptionAllowSteal = "allow_steal"
)

// Settings configure one game; which fields matter depends on the mode
type Settings struct {
	Count      int  `json:"count,omitempty"`       // words or sentences
	TimeLimit  int  `json:"time_limit,omitempty"`  // seconds; game length for Word Pool
	SpawnRate  int  `json:"spawn_rate,omitempty"`  // Word Pool: words per second
	AllowSteal bool `json:"allow_steal,omitempty"` // Word Pool: may take others' words
}

// Option is one adjustable setting of a mode. Values are the allowed steps,
// in wheel order; booleans use 0 for off and 1 for on.
type Option struct {
	Key     string
	Label   string
	Values  []int
	Default int
	Unit    string // appended to numbers, e.g. "s"
	Toggle  bool   // shown as Off/On
}

// Format renders one value of the option for display
func (o Option) Format(v int) string {
	if o.Toggle {
		if v != 0 {
			return "On"
		}
		return "Off"
	}
	return strconv.Itoa(v) + o.Unit
}

// Index returns the position of v among the allowed values, or -1
func (o Option) Index(v int) int {
	for i, allowed := range o.Values {
		if allowed == v {
			return i
		}
	}
	return -1
}

var modeOptions = map[string][]Option{
	ModeWordDrop: {
		{Key: OptionCount, Label: "Words", Values: []int{5, 10, 15, 20, 30}, Default: 10},
		{Key: OptionTimeLimit, Label: "Time limit", Values: []int{60, 90, 120, 180, 300}, Default: 120, Unit: "s"},
	},
	ModeSentenceRace: {
		{Key: OptionCount, Label: "Sentences", Values: []int{3, 5, 8, 10}, Default: 5},
		{Key: OptionTimeLimit, Label: "Time limit", Values: []int{60, 120, 180, 300}, Default: 180, Unit: "s"},
	},
	ModeWordPool: {
		{Key: OptionTimeLimit, Label: "Duration", Values: []int{30, 60, 90, 120}, Default: 60, Unit: "s"},
		{Key: OptionSpawnRate, Label: "Words/sec", Values: []int{1, 2, 3}, Default: 1},
		{Key: OptionAllowSteal, Label: "Steal", Values: []int{0, 1}, Default: 0, Toggle: true},
	},
}

// ModeOptions returns the adjustable settings of a mode, in menu order
func ModeOptions(mode string) []Option {
	return modeOptions[mode]
}

// DefaultSettings returns the defaults for a mode
func DefaultSettings(mode string) Settings {
	var s Settings
	for _, o := range modeOptions[mode] {
		s = s.with(o.Key, o.Default)
	}
	return s
}

// Get returns the value of a setting, booleans as 0 or 1
func (s Settings) Get(key string) int {
	switch key {
	case OptionCount:
		return s.Count
	case OptionTimeLimit:
		return s.TimeLimit
	case OptionSpawnRate:
		return s.SpawnRate
	case OptionAllowSteal:
		if s.AllowSteal {
			return 1
		}
	}
	return 0
}

func (s Settings) with(key string, v int) Settings {
	switch key {
	case OptionCount:
		s.Count = v
	case OptionTimeLimit:
		s.TimeLimit = v
	case OptionSpawnRate:
		s.SpawnRate = v
	case OptionAllowSteal:
		s.AllowSteal = v != 0
	}
	return s
}

// Apply returns the settings with one option changed, or false when the
// mode has no such option or the value is not one of its steps
func (s Settings) Apply(mode, key string, v int) (Settings, bool) {
	for _, o := range modeOptions[mode] {
		if o.Key == key && o.Index(v) >= 0 {
			return s.with(key, v), true
		}
	}
	return s, false
}

func isKnownMode(mode string) bool {
	_, ok := modeOptions[mode]
	return ok
}
