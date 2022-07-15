/* See LICENSE for license details. */
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

type OutFmt struct {
	ruby  string /* Ruby output format */
	hdr   string /* Header format */
	shdr  string /* Subheader format */
	sshdr string /* Subsubheader format */
	pb    string /* Page Break text */
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [-d] [-f format] file\n", os.Args[0])
	flag.PrintDefaults()
}

func get_outfmt(fmt string) *OutFmt {
	of := new(OutFmt)

	switch fmt {
	case "tex":
		of.ruby  = "\\ruby{%s}{%s}"
		of.hdr   = "\\chapter{%s}"
		of.shdr  = "\\section*{%s}"
		of.sshdr = "\\subsection*{%s}"
		of.pb    = "\\newpage"
	case "md":
		of.ruby  = "<ruby>%s<rp>《</rp><rt>%s</rt><rp>》</rp></ruby>"
		of.hdr   = "# %s"
		of.shdr  = "## %s"
		of.sshdr = "### %s"
		of.pb    = "<div style='break-after:always'></div>"
	case "plain":
		of.ruby  = "[%s:%s]"
		of.hdr   = "%s"
		of.shdr  = "%s"
		of.sshdr = "%s"
		of.pb    = ""
	}

	return of
}

func accent_map() map[string]string {
	/* https://web.archive.org/web/20220206093806/http://aozora.gr.jp/accent_separation.html */
	return map[string]string {
		"A&": "Å",
		"A'": "Á",
		"A:": "Ä",
		"AE&": "Æ",
		"A^": "Â",
		"A_": "Ā",
		"A`": "À",
		"A~": "Ã",
		"C'": "Ć",
		"C,": "Ç",
		"C^": "Ĉ",
		"D/": "Đ",
		"E'": "É",
		"E:": "Ë",
		"E^": "Ê",
		"E_": "Ē",
		"E`": "È",
		"E~": "Ẽ",
		"G^": "Ĝ",
		"H^": "Ĥ",
		"I'": "Í",
		"I:": "Ï",
		"I^": "Î",
		"I_": "Ī",
		"I`": "Ì",
		"I~": "Ĩ",
		"J^": "Ĵ",
		"L'": "Ĺ",
		"L/": "Ł",
		"M'": "Ḿ",
		"N'": "Ń",
		"N`": "Ǹ",
		"N~": "Ñ",
		"O'": "Ó",
		"O/": "Ø",
		"O:": "Ö",
		"OE&": "Œ",
		"O^": "Ô",
		"O_": "Ō",
		"O`": "Ò",
		"O~": "Õ",
		"R'": "Ŕ",
		"S'": "Ś",
		"S,": "Ş",
		"S^": "Ŝ",
		"T,": "Ţ",
		"U&": "Ů",
		"U'": "Ú",
		"U:": "Ü",
		"U^": "Û",
		"U_": "Ū",
		"U`": "Ù",
		"U~": "Ũ",
		"Y'": "Ý",
		"Z'": "Ź",
		"a&": "å",
		"a'": "á",
		"a:": "ä",
		"a^": "â",
		"a_": "ā",
		"a`": "à",
		"ae&": "æ",
		"a~": "ã",
		"c'": "ć",
		"c,": "ç",
		"c^": "ĉ",
		"d/": "đ",
		"e'": "é",
		"e:": "ë",
		"e^": "ê",
		"e_": "ē",
		"e`": "è",
		"e~": "ẽ",
		"g^": "ĝ",
		"h/": "ħ",
		"h^": "ĥ",
		"i'": "í",
		"i/": "ɨ",
		"i:": "ï",
		"i^": "î",
		"i_": "ī",
		"i`": "ì",
		"i~": "ĩ",
		"j^": "ĵ",
		"l'": "ĺ",
		"l/": "ł",
		"m'": "ḿ",
		"n'": "ń",
		"n`": "ǹ",
		"n~": "ñ",
		"o'": "ó",
		"o/": "ø",
		"o:": "ö",
		"o^": "ô",
		"o_": "ō",
		"o`": "ò",
		"oe&": "œ",
		"o~": "õ",
		"r'": "ŕ",
		"s&": "ß",
		"s'": "ś",
		"s,": "ş",
		"s^": "ŝ",
		"t,": "ţ",
		"u&": "ů",
		"u'": "ú",
		"u:": "ü",
		"u^": "û",
		"u_": "ū",
		"u`": "ù",
		"u~": "ũ",
		"y'": "ý",
		"y:": "ÿ",
		"z'": "ź",
	}
}

func jis_map() map[int]string {
	/* https://kanji.jitenon.jp/ */
	/* http://www13.plala.or.jp/bigdata/index_kanji.html */
	return map[int]string {
		311476: "匇",
		311524: "噱",
		311589: "媧",
		318428: "彘",
		318431: "彽",
		318445: "怳",
		318454: "惝",
		318455: "惸",
		318459: "愷",
		318466: "戢",
		318477: "挘",
		318615: "橛",
		318662: "泫",
		318740: "炷",
		318764: "燄",
		318771: "犍",
		318822: "璆",
		318881: "眶",
		318885: "睜",
		319155: "蛼",
		319239: "蹰",
		319278: "鄢",
		319413: "騃",
		319484: "鼹",
		421283: "戕",
		428874: "譃",
		429267: "餼",
		429268: "饀",
		429271: "饍",
		429337: "魳",
	}
}

func replace_jis(str string) string {
	exp := regexp.MustCompile(`※［＃([^］]+)］`)

	for _, matches := range exp.FindAllStringSubmatch(str, -1) {
		sub_exp := regexp.MustCompile(`第(\d)水準(\d)-(\d\d)-(\d\d)`)
	
		nums := sub_exp.FindStringSubmatch(str)
		if nums == nil {
			/* the same character appeared multiple times in str */
			continue
		}
		num, _ := strconv.Atoi(nums[1] + nums[2] + nums[3] + nums[4])

		m := jis_map()
		replacement, ok := m[num]
		if !ok {
			log.Printf("jis code not implemented: %d: %s\n", num, matches[0])
			continue
		}

		str = strings.Replace(str, matches[0], replacement, -1)
	}

	return str
}

func replace_ruby(str string, of *OutFmt) string {
	kanji := `\x{3400}-\x{4DBF}` +   /* CJK Unified Ideographs Extension A */
		 `\x{4E00}-\x{9FFF}` +   /* CJK Unified Ideographs */
		 `\x{F900}-\x{FAFF}` +   /* CJK Compatibility Ideographs */
		 `\x{20000}-\x{2FA1F}` + /* CJK Unified Ideographs Extension B - F, Supplement */
		 `〆〻〇々ヶ`
	ruby_exp := regexp.MustCompile(`[｜]?([` + kanji + `]+)《([^》]+)》`)
	for _, matches := range ruby_exp.FindAllStringSubmatch(str, -1) {
		replacement := fmt.Sprintf(of.ruby, matches[1], matches[2])
		str = strings.Replace(str, matches[0], replacement, -1)
	}

	bouten_exp := regexp.MustCompile(`［＃「([^」]+)」に傍点］`)
	for _, matches := range bouten_exp.FindAllStringSubmatch(str, -1) {
		bouten := strings.Repeat("﹅", utf8.RuneCountInString(matches[1]))
		replacement := fmt.Sprintf(of.ruby, matches[1], bouten)
		str = strings.Replace(str, matches[1] + matches[0], replacement, -1)
	}

	return str
}

func replace_accents(str string) string {
	exp := regexp.MustCompile(`〔([^〕]+)〕`)
	
	for _, matches := range exp.FindAllStringSubmatch(str, -1) {
		str = strings.Replace(str, matches[0], matches[1], -1)

		m := accent_map()
		for key := range m {
			str = strings.ReplaceAll(str, key, m[key])
		}
	}

	return str
}

func replace_hdrs(str string, of *OutFmt) string {
	exp := regexp.MustCompile(`\n\n［[^［]+［＃「([^」]+)」は([大中小])見出し］\n\n\n`)
	slices := exp.FindAllStringSubmatch(str, -1)
	if slices == nil {
		exp = regexp.MustCompile(`\n\n\n([^\n]+)\n\n\n`)
		for _, matches := range exp.FindAllStringSubmatch(str, -1) {
			replacement := "\n" + fmt.Sprintf(of.hdr, matches[1]) + "\n"
			str = strings.Replace(str, matches[0], replacement, -1)
		}
		return str
	}

	for _, matches := range slices {
		var replacement string
		switch matches[2] {
		case "大":
			replacement = fmt.Sprintf(of.hdr, matches[1])
		case "中":
			replacement = fmt.Sprintf(of.shdr, matches[1])
		case "小":
			replacement = fmt.Sprintf(of.sshdr, matches[1])
		default:
			log.Printf("bad hdr: %s\n", matches[0])
			replacement = matches[1]
		}
		str = strings.Replace(str, matches[0], replacement + "\n", -1)
	}

	return str
}

func trim_info(str string) string {
	delim := "\n" + strings.Repeat("-", 55) + "\n"

	slices := strings.Split(str, delim)

	return strings.Join([]string{slices[0], slices[2]}, "")
}

func parse(file string, of *OutFmt, debug bool) string {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}

	var lines []string
	r := bufio.NewScanner(f)
	for r.Scan() {
		line := strings.Trim(r.Text(), "　")
		line = replace_jis(line)
		line = replace_ruby(line, of)
		line = replace_accents(line)
		lines = append(lines, line)
	}

	out := strings.Join(lines, "\n\n");
	out = replace_hdrs(out, of)
	out = strings.Replace(out, "［＃改ページ］", of.pb, -1)

	if (debug == false) {
		out = trim_info(out)
	}

	return out
}

func main() {
	var (
		debug = flag.Bool("d", false, "debug mode")
		format = flag.String("f", "plain", "output format [plain|md|tex]")
	)

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
		os.Exit(1)
	}

	log.SetFlags(log.Lshortfile)

	of := get_outfmt(*format)
	out := parse(flag.Arg(0), of, *debug) 

	fmt.Printf("%s\n", out)
}
