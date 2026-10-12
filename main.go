package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// ---------- simulatie (zelfde regels als cpu.go) ----------

var elementsDepthArray [256][4]byte
var moveToIndexArray [256]int

func init() {
	elementsDepthArray['W'] = [4]byte{'L', 'A', 'V', 'W'}
	elementsDepthArray['V'] = [4]byte{'W', 'L', 'A', 'V'}
	elementsDepthArray['A'] = [4]byte{'V', 'W', 'L', 'A'}
	elementsDepthArray['L'] = [4]byte{'A', 'V', 'W', 'L'}
	moveToIndexArray['W'] = 0
	moveToIndexArray['V'] = 1
	moveToIndexArray['A'] = 2
	moveToIndexArray['L'] = 3
	moveToIndexArray['D'] = 4
}

var moveWins = [5][5]uint8{
	{0, 1, 0, 2, 0},
	{2, 0, 1, 0, 0},
	{0, 2, 0, 1, 0},
	{1, 0, 2, 0, 0},
	{0, 0, 0, 0, 0},
}

var depthToElement = [5]byte{'W', 'V', 'A', 'L', 'D'}

type Player struct {
	available [5]int
	moves     [13]byte
	moveCount int
}

func getElementFromCode(depth int) byte {
	if depth < 1 || depth > 5 {
		return 0
	}
	return depthToElement[depth-1]
}

func getElementByDepth(prev byte, depth int) byte {
	if depth == 5 {
		return 'D'
	}
	if prev == 0 {
		return 0
	}
	if prev == 'D' {
		prev = 'L'
	}
	return elementsDepthArray[prev][depth-1]
}

func chooseAvailableElement(target byte, available *[5]int) byte {
	if available[moveToIndexArray[target]] > 0 {
		return target
	}
	current := target
	for i := 0; i < 5; i++ {
		current = elementsDepthArray[current][0]
		if available[moveToIndexArray[current]] > 0 {
			return current
		}
	}
	if available[4] > 0 {
		return 'D'
	}
	return 0
}

func getLastElement(available *[5]int) byte {
	for i, c := range depthToElement {
		if available[i] > 0 {
			return c
		}
	}
	return 0
}

func determineWinner(m1, m2 byte) int {
	return int(moveWins[moveToIndexArray[m1]][moveToIndexArray[m2]])
}

// resolveTarget bepaalt het gewenste element voor zet i: een letter (WVALD) is een vaste zet, een cijfer een diepte
// en de zetten van de tegenstander.
func resolveTarget(ch byte, i int, oppMoves []byte) byte {
	if isMoveLetter(ch) {
		return ch // vaste zet (letter)
	}
	digit := int(ch - '0')
	if digit >= 6 {
		base := digit - 5
		if i < 2 {
			return getElementFromCode(base)
		}
		return getElementByDepth(oppMoves[i-2], base)
	}
	if i == 0 {
		return getElementFromCode(digit)
	}
	return getElementByDepth(oppMoves[i-1], digit)
}

// simulateDepthGame: twee diepte-engines (12 cijfers) tegen elkaar, volledige score.
func simulateDepthGame(e1, e2 string) (s1, s2 int) {
	var p1, p2 Player
	p1.available = [5]int{3, 3, 3, 3, 1}
	p2.available = [5]int{3, 3, 3, 3, 1}
	for i := 0; i < 12; i++ {
		t1 := resolveTarget(e1[i], i, p2.moves[:])
		t2 := resolveTarget(e2[i], i, p1.moves[:])
		m1 := chooseAvailableElement(t1, &p1.available)
		m2 := chooseAvailableElement(t2, &p2.available)
		if m1 == 0 || m2 == 0 {
			return -1, -1
		}
		p1.available[moveToIndexArray[m1]]--
		p1.moves[i] = m1
		p2.available[moveToIndexArray[m2]]--
		p2.moves[i] = m2
		switch determineWinner(m1, m2) {
		case 1:
			s1++
		case 2:
			s2++
		}
	}
	m1 := getLastElement(&p1.available)
	m2 := getLastElement(&p2.available)
	switch determineWinner(m1, m2) {
	case 1:
		s1++
	case 2:
		s2++
	}
	return s1, s2
}

// simulateDepthVsFixed: diepte-engine reageert op een vaste 13-zet engine.
func simulateDepthVsFixed(engine, fixed string) (s1, s2 int) {
	p := Player{available: [5]int{3, 3, 3, 3, 1}}
	var moves [13]byte
	for i := 0; i < 12; i++ {
		t := resolveTarget(engine[i], i, []byte(fixed))
		m := chooseAvailableElement(t, &p.available)
		if m == 0 {
			m = getLastElement(&p.available)
		}
		if m != 0 {
			p.available[moveToIndexArray[m]]--
		}
		moves[i] = m
	}
	m := getLastElement(&p.available)
	if m == 0 {
		m = 'W'
	}
	moves[12] = m
	for i := 0; i < 13; i++ {
		switch determineWinner(moves[i], fixed[i]) {
		case 1:
			s1++
		case 2:
			s2++
		}
	}
	return s1, s2
}

// simulateFixedGame: twee vaste engines (13 zetten) tegen elkaar.
func simulateFixedGame(e1, e2 string) (s1, s2 int) {
	for i := 0; i < 13; i++ {
		switch determineWinner(e1[i], e2[i]) {
		case 1:
			s1++
		case 2:
			s2++
		}
	}
	return s1, s2
}

// play speelt engine a tegen engine b (elk 12-cijferige diepte of 13-tekens vast) en
// geeft de punten van a: 1 bij winst, 0,5 bij gelijkspel, 0 bij verlies.
func play(a, b string) float64 {
	var sa, sb int
	switch {
	case len(a) == 12 && len(b) == 12:
		sa, sb = simulateDepthGame(a, b)
	case len(a) == 12 && len(b) == 13:
		sa, sb = simulateDepthVsFixed(a, b)
	case len(a) == 13 && len(b) == 12:
		sb, sa = simulateDepthVsFixed(b, a)
	default:
		sa, sb = simulateFixedGame(a, b)
	}
	switch {
	case sa > sb:
		return 1
	case sa == sb:
		return 0.5
	}
	return 0
}

func parseEngineCode(input string) string {
	parts := strings.Split(input, ":")
	engine := strings.TrimSpace(input)
	if len(parts) > 2 {
		engine = strings.TrimSpace(parts[2])
	}
	if isFixed13 := len(engine) == 13 && strings.Trim(engine, "WVALD") == ""; len(engine) > 12 && !isFixed13 {
		return engine[:12]
	}
	return engine
}

func validCode(e string) bool {
	return isValidDepthCode(e) ||
		(len(e) == 13 && strings.Trim(e, "WVALD") == "")
}

type levelEngine struct {
	level int
	code  string
}

func parseProtocolLine(line string) (levelEngine, error) {
	parts := strings.Split(line, ":")
	if len(parts) < 3 {
		return levelEngine{}, fmt.Errorf("formaat moet Naam:level:code zijn")
	}
	lv, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || lv < 1 {
		return levelEngine{}, fmt.Errorf("ongeldig level '%s'", strings.TrimSpace(parts[1]))
	}
	e := parseEngineCode(line)
	if !validCode(e) {
		return levelEngine{}, fmt.Errorf("ongeldige engine code '%s'", e)
	}
	return levelEngine{level: lv, code: e}, nil
}

// minRatio is de minimale score (fractie van de punten) voor "voldoende" tegen een level.
const minRatio = 0.7

func fmtPoints(p float64) string {
	return strconv.FormatFloat(p, 'f', -1, 64)
}

func main() {
	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)

	for {
		fmt.Print("Voer de engine in die moet spelen (code of Naam:level:code, leeg om te stoppen): ")
		if !sc.Scan() {
			return
		}
		line := strings.TrimSpace(sc.Text())
		if line == "" || line == "." {
			fmt.Println("Gestopt.")
			return
		}
		player := parseEngineCode(line)
		if !validCode(player) {
			fmt.Printf("Ongeldige engine code '%s'. Moet 12 cijfers (1-9) of 13 tekens (W, V, A, L, D) zijn.\n\n", player)
			continue
		}

		fmt.Println("Voer alle engines in met formaat Naam:level:code (één per regel, '.' of lege regel om te stoppen):")
		var all []levelEngine
		for sc.Scan() {
			l := strings.TrimSpace(sc.Text())
			if l == "" || l == "." {
				break
			}
			e, err := parseProtocolLine(l)
			if err != nil {
				fmt.Printf("Regel '%s' overgeslagen: %v\n", l, err)
				continue
			}
			all = append(all, e)
		}
		if len(all) == 0 {
			fmt.Print("Geen engines ingevoerd.\n\n")
			continue
		}

		// punten en aantal tegenstanders per level
		points := map[int]float64{}
		count := map[int]int{}
		for _, e := range all {
			points[e.level] += play(player, e.code)
			count[e.level]++
		}
		levels := make([]int, 0, len(points))
		for lv := range points {
			levels = append(levels, lv)
		}
		// hoogste score eerst; bij gelijke score het hoogste level eerst
		sort.Slice(levels, func(i, j int) bool {
			a, b := levels[i], levels[j]
			if points[a] != points[b] {
				return points[a] > points[b]
			}
			return a > b
		})

		fmt.Printf("\nEngine %s tegen %d engines in %d levels:\n", player, len(all), len(levels))
		var total float64
		for _, lv := range levels {
			fmt.Printf("Level %d -> %s/%d\n", lv, fmtPoints(points[lv]), count[lv])
			total += points[lv]
		}
		fmt.Printf("Totaal: %s/%d\n", fmtPoints(total), len(all))

		// hoogste level waar minstens 70% van de punten gehaald is
		best := 0
		for lv, p := range points {
			if p >= minRatio*float64(count[lv])-1e-9 && lv > best {
				best = lv
			}
		}
		if best > 0 {
			fmt.Printf("Hoogste level met %.0f%% of meer: Level %d (%s/%d)\n\n", minRatio*100, best, fmtPoints(points[best]), count[best])
		} else {
			fmt.Printf("Geen enkel level met %.0f%% of meer.\n\n", minRatio*100)
		}
	}
}

func isMoveLetter(c byte) bool {
	return c == 'W' || c == 'V' || c == 'A' || c == 'L' || c == 'D'
}

// isValidDepthCode aanvaardt 12 tekens: eerst letters (WVALD, vaste zet), dan cijfers (1-9, diepte).
// bv. AVAWD6161164: speelt eerst A,V,A,W,D en daarna de dieptes 6,1,6,1,1,6,4.
func isValidDepthCode(engine string) bool {
	if len(engine) != 12 {
		return false
	}
	seenDigit := false
	for i := 0; i < len(engine); i++ {
		c := engine[i]
		switch {
		case c >= '1' && c <= '9':
			seenDigit = true
		case isMoveLetter(c):
			if seenDigit {
				return false
			}
		default:
			return false
		}
	}
	return true
}
