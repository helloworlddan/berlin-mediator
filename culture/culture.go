// Warning: pseudo science
package culture

import (
	"errors"
	"fmt"
	"strings"
)

type Style struct {
	Communication int
	Evaluation    int
	Leading       int
	Deciding      int
	Trusting      int
	Disagreeing   int
	Scheduling    int
}

func (s Style) TextIntensities() []string {
	intensityMap := map[int]string{
		-10: "WAY LESS",
		-9:  "WAY LESS",
		-8:  "way less",
		-7:  "way less",
		-6:  "a lot less",
		-5:  "a lot less",
		-4:  "less",
		-3:  "less",
		-2:  "slightly less",
		-1:  "slightly less",
		0:   "",
		1:   "slightly more",
		2:   "slightly more",
		3:   "more",
		4:   "more",
		5:   "a lot more",
		6:   "a lot more",
		7:   "way more",
		8:   "way more",
		9:   "WAY MORE",
		10:  "WAY MORE",
	}

	texts := []string{}

	texts = append(texts, fmt.Sprintf("%s communicative", intensityMap[s.Communication]))
	texts = append(texts, fmt.Sprintf("%s evaluating", intensityMap[s.Evaluation]))
	texts = append(texts, fmt.Sprintf("%s leading", intensityMap[s.Leading]))
	texts = append(texts, fmt.Sprintf("%s decisive", intensityMap[s.Deciding]))
	texts = append(texts, fmt.Sprintf("%s trusting", intensityMap[s.Trusting]))
	texts = append(texts, fmt.Sprintf("%s disagreeing", intensityMap[s.Disagreeing]))
	texts = append(texts, fmt.Sprintf("%s flexible", intensityMap[s.Scheduling]))

	return texts
}

func Delta(source string, target string) (Style, error) {
	cultureMap := map[string]Style{
		"de": {
			Communication: -2,
			Evaluation:    -3,
			Leading:       0,
			Deciding:      -1,
			Trusting:      -2,
			Disagreeing:   -2,
			Scheduling:    -4,
		},
		"tr": {
			Communication: 3,
			Evaluation:    1,
			Leading:       3,
			Deciding:      3,
			Trusting:      3,
			Disagreeing:   1,
			Scheduling:    3,
		},
		"pl": {
			Communication: 0,
			Evaluation:    -2,
			Leading:       3,
			Deciding:      3,
			Trusting:      0,
			Disagreeing:   1,
			Scheduling:    0,
		},
		"ar": {
			Communication: 3,
			Evaluation:    1,
			Leading:       4,
			Deciding:      3,
			Trusting:      3,
			Disagreeing:   2,
			Scheduling:    4,
		},
		"it": {
			Communication: 2,
			Evaluation:    -1,
			Leading:       3,
			Deciding:      4,
			Trusting:      2,
			Disagreeing:   -1,
			Scheduling:    1,
		},
	}

	sourceStyle, ok := cultureMap[strings.ToLower(source)]
	if !ok {
		return Style{}, errors.New("source culture not found")
	}

	targetStyle, ok := cultureMap[strings.ToLower(target)]
	if !ok {
		return Style{}, errors.New("target culture not found")
	}

	delta := Style{
		Communication: distance(sourceStyle.Communication, targetStyle.Communication),
		Evaluation:    distance(sourceStyle.Evaluation, targetStyle.Evaluation),
		Leading:       distance(sourceStyle.Leading, targetStyle.Leading),
		Deciding:      distance(sourceStyle.Deciding, targetStyle.Deciding),
		Trusting:      distance(sourceStyle.Trusting, targetStyle.Trusting),
		Disagreeing:   distance(sourceStyle.Disagreeing, targetStyle.Disagreeing),
		Scheduling:    distance(sourceStyle.Scheduling, targetStyle.Scheduling),
	}
	return delta, nil
}

func distance(a int, b int) int {
	diffAbs := abs(a - b)

	if a > b {
		return diffAbs * -1
	}

	return diffAbs
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
