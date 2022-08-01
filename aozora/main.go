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

	"aozora2fmt"
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

		m := aozora2fmt.JisMap()
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

		m := aozora2fmt.AccentMap()
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
