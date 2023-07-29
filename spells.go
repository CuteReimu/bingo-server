package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

// RandSpells 随符卡，用于标准模式
func RandSpells(games []string, ranks []string, lvCount [3]int) ([]*Spell, error) {
	idx := []int{0, 1, 3, 4}
	rand.Shuffle(len(idx), func(i, j int) { idx[i], idx[j] = idx[j], idx[i] })
	star123 := make([]int32, 0, 20)
	for i := 0; i < lvCount[0]; i++ {
		star123 = append(star123, 1)
	}
	for i := 0; i < lvCount[1]; i++ {
		star123 = append(star123, 2)
	}
	for i := 0; i < lvCount[2]; i++ {
		star123 = append(star123, 3)
	}
	rand.Shuffle(len(star123), func(i, j int) { star123[i], star123[j] = star123[j], star123[i] })
	star45 := []int32{4, 4, 4, 4, 5}
	rand.Shuffle(len(star45), func(i, j int) { star45[i], star45[j] = star45[j], star45[i] })
	j := 0
	stars := make([]int32, 0, 25)
	for i := 0; i < 25; i++ {
		switch i {
		// 每行、每列都只有一个大于等于lv4
		case idx[0]:
			stars = append(stars, star45[0])
		case 5 + idx[1]:
			stars = append(stars, star45[1])
		case 12:
			stars = append(stars, star45[2])
		case 15 + idx[2]:
			stars = append(stars, star45[3])
		case 20 + idx[3]:
			stars = append(stars, star45[4])
		default:
			stars = append(stars, star123[j])
			j++
		}
	}
	return NormalGameSpellConfig.Get(games, ranks, ranksToExPos(ranks), stars)
}

// RandSpellsLink 随符卡，用于link赛
func RandSpellsLink(games []string, ranks []string, lvCount [3]int) ([]*Spell, error) {
	idx := []int{0, 1, 3, 4}
	rand.Shuffle(len(idx), func(i, j int) { idx[i], idx[j] = idx[j], idx[i] })
	star123 := make([]int32, 0, 20)
	for i := 0; i < lvCount[0]; i++ {
		star123 = append(star123, 1)
	}
	for i := 0; i < lvCount[1]; i++ {
		star123 = append(star123, 2)
	}
	for i := 0; i < lvCount[2]; i++ {
		star123 = append(star123, 3)
	}
	rand.Shuffle(len(star123), func(i, j int) { star123[i], star123[j] = star123[j], star123[i] })
	j := 0
	stars := make([]int32, 0, 25)
	for i := 0; i < 25; i++ {
		switch i {
		case 0, 4: // 左上lv1，右上lv1
			stars = append(stars, 1)
		case 6, 8, 16, 18: // 第二、四排的第二、四列固定4级
			stars = append(stars, 4)
		case 12: // 中间5级
			stars = append(stars, 5)
		default:
			stars = append(stars, star123[j])
			j++
		}
	}
	return NormalGameSpellConfig.Get(games, ranks, ranksToExPos(ranks), stars)
}

func ranksToExPos(ranks []string) []int {
	if len(ranks) > 0 && !slices.ContainsFunc(ranks, func(s string) bool { return s != "L" }) {
		return nil
	}
	idx := []int{0, 1, 2, 3, 4}
	rand.Shuffle(len(idx), func(i, j int) { idx[i], idx[j] = idx[j], idx[i] })
	for i, j := range idx {
		idx[i] = i*5 + j
	}
	return idx
}

func (x SpellStatus) isSelectStatus() bool {
	return x == SpellStatus_left_select || x == SpellStatus_right_select || x == SpellStatus_both_select
}

func (x SpellStatus) isGetStatus() bool {
	return x == SpellStatus_left_get || x == SpellStatus_right_get || x == SpellStatus_both_get
}

func (x SpellStatus) isLeftStatus() bool {
	return x == SpellStatus_left_select || x == SpellStatus_left_get || x == SpellStatus_both_select || x == SpellStatus_both_get
}

func (x SpellStatus) isRightStatus() bool {
	return x == SpellStatus_right_select || x == SpellStatus_right_get || x == SpellStatus_both_select || x == SpellStatus_both_get
}

type SpellConfig struct {
	SpellBuilder func([]string) *Spell
	md5Sum       map[string]bool
	allSpells    map[int32]map[bool]map[string][]*Spell // star => ( isEx => ( game => spellList ) )
}

func (c *SpellConfig) Get(games, ranks []string, exPos []int, stars []int32) ([]*Spell, error) {
	m := make(map[int32]map[bool]map[string][]*Spell)
	filteredMap, err := c.getOrLoad()
	if err != nil {
		return nil, err
	}
	for star, isExMap := range filteredMap {
		isExMap2 := make(map[bool]map[string][]*Spell)
		for isEx, gameMap := range isExMap {
			for game, spellList := range gameMap {
				if !slices.Contains(games, game) {
					continue
				}
				var spellList2 []*Spell
				if len(ranks) > 0 {
					spellList2 = nil
					for _, spell := range spellList {
						if slices.Contains(ranks, spell.Rank) {
							spellList2 = append(spellList2, spell)
						}
					}
				} else {
					spellList2 = slices.Clone(spellList)
				}
				if len(spellList2) > 0 {
					gameMap2, ok := isExMap2[isEx]
					if !ok {
						gameMap2 = make(map[string][]*Spell)
						isExMap2[isEx] = gameMap2
					}
					rand.Shuffle(len(spellList2), func(i, j int) {
						spellList2[i], spellList2[j] = spellList2[j], spellList2[i]
					})
					gameMap2[game] = spellList2
				}
			}
		}
		if len(isExMap2) > 0 {
			m[star] = isExMap2
		}
	}
	spellIds := make(map[string]bool)
	result := make([]*Spell, 0, len(stars))
	for i := range stars {
		if isExMap, ok := m[stars[i]]; !ok {
			return nil, errors.New("符卡数量不足")
		} else if gameMap, ok := isExMap[false]; !ok {
			return nil, errors.New("符卡数量不足")
		} else {
			var spell *Spell
			for {
				games := maps.Keys(gameMap)
				if len(games) == 0 {
					return nil, errors.New("符卡数量不足")
				}
				game := games[rand.Intn(len(games))]
				spellList := gameMap[game]
				spell = spellList[0]
				spellList = spellList[1:]
				if len(spellList) == 0 {
					delete(gameMap, game)
				} else {
					gameMap[game] = spellList
				}
				spellId := fmt.Sprintf("%s-%d", spell.Game, spell.Id)
				if !spellIds[spellId] {
					spellIds[spellId] = true
					break
				}
			}
			result = append(result, spell)
		}
	}
	for i := range exPos {
		index := exPos[i]
		firstTry := true
	tryOnce:
		for {
			if firstTry {
				firstTry = false
			} else {
				index = (index + 1) % len(result)
				if index == exPos[i] {
					return nil, errors.New("符卡数量不足")
				}
				if slices.Contains(exPos, index) {
					continue
				}
			}
			if isExMap, ok := m[stars[index]]; !ok {
				continue
			} else if gameMap, ok := isExMap[true]; !ok {
				continue
			} else {
				var spell *Spell
				for {
					games := maps.Keys(gameMap)
					if len(games) == 0 {
						continue tryOnce
					}
					game := games[rand.Intn(len(games))]
					spellList := gameMap[game]
					spell = spellList[0]
					spellList = spellList[1:]
					if len(spellList) == 0 {
						delete(gameMap, game)
					} else {
						gameMap[game] = spellList
					}
					spellId := fmt.Sprintf("%s-%d", spell.Game, spell.Id)
					if !spellIds[spellId] {
						spellIds[spellId] = true
						break
					}
				}
				exPos[i] = index
				result[index] = spell
				break
			}
		}
	}
	return result, nil
}

func (c *SpellConfig) getOrLoad() (map[int32]map[bool]map[string][]*Spell, error) {
	dirEntries, err := os.ReadDir(".")
	if err != nil {
		return nil, errors.Wrap(err, "找不到符卡文件")
	}
	var files []string
	for _, dirEntry := range dirEntries {
		name := dirEntry.Name()
		if strings.HasSuffix(name, ".xlsx") && !strings.HasPrefix(name, "log") {
			files = append(files, name)
		}
	}

	md5SumFunc := func(fileName string) string {
		file, err := os.Open(fileName)
		if err != nil {
			return ""
		}
		defer func() { _ = file.Close() }()
		hash := md5.New()
		if _, err := io.Copy(hash, file); err != nil {
			return ""
		}
		return hex.EncodeToString(hash.Sum(nil))
	}

	md5Sum := make(map[string]bool)
	for _, file := range files {
		md5Sum[md5SumFunc(file)] = true
	}
	if maps.Equal(md5Sum, c.md5Sum) {
		return c.allSpells, nil
	}
	allSpells := make(map[int32]map[bool]map[string][]*Spell)
	for _, file := range files {
		xlsx, err := excelize.OpenFile(file)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		rows, err := xlsx.GetRows("Sheet1")
		if err != nil {
			return nil, errors.WithStack(err)
		}
		for i, row := range rows {
			if i > 0 {
				spell := c.SpellBuilder(row)
				if spell != nil {
					isExMap, ok := allSpells[spell.Star]
					if !ok {
						isExMap = make(map[bool]map[string][]*Spell)
						allSpells[spell.Star] = isExMap
					}
					gameMap, ok := isExMap[spell.Rank != "L"]
					if !ok {
						gameMap = make(map[string][]*Spell)
						isExMap[spell.Rank != "L"] = gameMap
					}
					gameMap[spell.Game] = append(gameMap[spell.Game], spell)
				}
			}
		}
	}
	c.md5Sum = md5Sum
	c.allSpells = allSpells
	return allSpells, nil
}

var (
	NormalGameSpellConfig = &SpellConfig{SpellBuilder: func(row []string) *Spell {
		if len(row) < 7 {
			return nil
		}
		star, err := strconv.Atoi(row[6])
		if err != nil {
			return nil
		}
		var id int
		if len(row) > 8 {
			id, _ = strconv.Atoi(row[8])
		}
		return &Spell{Game: row[1], Name: row[3], Rank: row[5], Star: int32(star), Desc: row[4], Id: int32(id)}
	}}

	//BPGameSpellConfig = &SpellConfig{SpellBuilder: func(row []string) *Spell {
	//	if len(row) < 7 {
	//		return nil
	//	}
	//	star, err := strconv.Atoi(row[7])
	//	if err != nil {
	//		return nil
	//	}
	//	var id int
	//	if len(row) > 8 {
	//		id, _ = strconv.Atoi(row[8])
	//	}
	//	return &Spell{Game: row[1], Name: row[3], Rank: row[5], Star: int32(star), Desc: row[4], Id: int32(id)}
	//}}
)
