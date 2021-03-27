package translate

import (
	"regexp"
)

type (
	Alphabet     string
	AlphabetList []Alphabet
)

const (
	JapaneseAlphabet Alphabet = "japanese"
	RussianAlphabet  Alphabet = "russian"
	LatinAlphabet    Alphabet = "latin"
	ArabicAlphabet   Alphabet = "arabic"
)

type L8n string

const (
	JapaneseL8n L8n = "ja"
	RussianL8n  L8n = "ru"
	EnglishL8n  L8n = "en"
	ArabicL8n   L8n = "ar"
)

var alphabetList = map[Alphabet]*regexp.Regexp{
	JapaneseAlphabet: regexp.MustCompile("[ぁ-んァ-ンｧ-ﾝﾞﾟぁ-ゞァ-ヶ]+"),
	RussianAlphabet:  regexp.MustCompile("[ЁёА-я]+"),
	LatinAlphabet:    regexp.MustCompile("[A-z]+"),
	ArabicAlphabet:   regexp.MustCompile("[ء-ي]+"),
}

func (as AlphabetList) Contains(alphabet Alphabet) (ok bool) {
	for _, a := range as {
		if a == alphabet {
			ok = true

			return
		}
	}

	return
}

func DetectAlphabets(line string) (list AlphabetList) {
	for a, re := range alphabetList {
		if re.MatchString(line) {
			list = append(list, a)
		}
	}

	return
}

func (a Alphabet) GetL8n() (l8n L8n) {
	switch a {
	case JapaneseAlphabet:
		l8n = JapaneseL8n
	case RussianAlphabet:
		l8n = RussianL8n
	case LatinAlphabet:
		l8n = EnglishL8n
	case ArabicAlphabet:
		l8n = ArabicL8n
	default:
		l8n = EnglishL8n
	}

	return
}

func DetectL8n(line string) (l8n L8n) {
	as := DetectAlphabets(line)

	if as.Contains(ArabicAlphabet) {
		return ArabicAlphabet.GetL8n()
	} else if as.Contains(JapaneseAlphabet) {
		return JapaneseAlphabet.GetL8n()
	} else if as.Contains(RussianAlphabet) {
		return RussianAlphabet.GetL8n()
	}

	return LatinAlphabet.GetL8n()
}
