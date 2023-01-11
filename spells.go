package main

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
	"golang.org/x/exp/slices"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func RandSpells(games []string, ranks []string, spellCounts [3]int) ([]*Spell, error) {
	if spellCounts[0]+spellCounts[1] != 20 || spellCounts[2] != 5 {
		panic(fmt.Sprint("错误的符卡数量", spellCounts[0]+spellCounts[1]+spellCounts[2]))
	}
	dirEntries, err := os.ReadDir(".")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	spells := make([][]*Spell, 4)
	for _, dirEntry := range dirEntries {
		fileName := dirEntry.Name()
		if strings.HasSuffix(fileName, ".xlsx") {
			xlsx, err := excelize.OpenFile(fileName)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			rows, err := xlsx.GetRows("Sheet1")
			if err != nil {
				return nil, errors.WithStack(err)
			}
			for i, row := range rows {
				if i > 0 && len(row) >= 7 {
					star, err := strconv.ParseInt(row[6], 10, 32)
					if err != nil {
						return nil, errors.WithStack(err)
					}
					inGame := slices.Contains(games, strings.TrimSpace(row[1])) && (ranks == nil || slices.Contains(ranks, strings.TrimSpace(row[5])))
					if star > 0 && star <= 3 && inGame {
						spells[star-1] = append(spells[star-1], &Spell{
							Game: row[1],
							Name: row[3],
							Rank: row[5],
							Star: int32(star),
							Desc: row[4],
						})
					}
					if star == 3 && !inGame {
						spells[3] = append(spells[3], &Spell{
							Game: row[1],
							Name: row[3],
							Rank: row[5],
							Star: int32(star),
							Desc: row[4],
						})
					}
				}
			}
		}
	}
	if len(spells[0]) < spellCounts[0] || len(spells[1]) < spellCounts[1] || len(spells[2])+len(spells[3]) < spellCounts[2] {
		return nil, errors.New("符卡数量不足")
	}
	r := rand.New(rand.NewSource(time.Now().UnixMilli()))
	r.Shuffle(len(spells[0]), func(i, j int) { spells[0][i], spells[0][j] = spells[0][j], spells[0][i] })
	r.Shuffle(len(spells[1]), func(i, j int) { spells[1][i], spells[1][j] = spells[1][j], spells[1][i] })
	spells01 := append(spells[0][:spellCounts[0]:spellCounts[0]], spells[1][:spellCounts[1]]...)
	r.Shuffle(len(spells01), func(i, j int) { spells01[i], spells01[j] = spells01[j], spells01[i] })
	if len(spells[2]) < spellCounts[2] {
		r.Shuffle(len(spells[3]), func(i, j int) { spells[3][i], spells[3][j] = spells[3][j], spells[3][i] })
		spells[2] = append(spells[2], spells[3][:spellCounts[2]-len(spells[2])]...)
	}
	r.Shuffle(len(spells[2]), func(i, j int) { spells[2][i], spells[2][j] = spells[2][j], spells[2][i] })
	idx := []int{0, 1, 3, 4}
	r.Shuffle(len(idx), func(i, j int) { idx[i], idx[j] = idx[j], idx[i] })
	result := make([]*Spell, 25)
	result[idx[0]] = spells[2][0]
	result[5+idx[1]] = spells[2][1]
	result[12] = spells[2][2]
	result[15+idx[2]] = spells[2][3]
	result[20+idx[3]] = spells[2][4]
	j := 0
	for i := range result {
		if result[i] == nil {
			result[i] = spells01[j]
			j++
		}
	}
	return result, nil
}

func RandSpells2(games []string, ranks []string, spellCounts [3]int) ([]*Spell, error) {
	if spellCounts[0]+spellCounts[1] != 20 || spellCounts[2] != 5 {
		panic(fmt.Sprint("错误的符卡数量", spellCounts[0]+spellCounts[1]+spellCounts[2]))
	}
	dirEntries, err := os.ReadDir(".")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	spells := make([][]*Spell, 4)
	for _, dirEntry := range dirEntries {
		fileName := dirEntry.Name()
		if strings.HasSuffix(fileName, ".xlsx") {
			xlsx, err := excelize.OpenFile(fileName)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			rows, err := xlsx.GetRows("Sheet1")
			if err != nil {
				return nil, errors.WithStack(err)
			}
			for i, row := range rows {
				if i > 0 && len(row) >= 7 {
					star, err := strconv.ParseInt(row[6], 10, 32)
					if err != nil {
						return nil, errors.WithStack(err)
					}
					inGame := slices.Contains(games, strings.TrimSpace(row[1])) && (ranks == nil || slices.Contains(ranks, strings.TrimSpace(row[5])))
					if star > 0 && star <= 3 && inGame {
						spells[star-1] = append(spells[star-1], &Spell{
							Game: row[1],
							Name: row[3],
							Rank: row[5],
							Star: int32(star),
							Desc: row[4],
						})
					}
					if star == 3 && !inGame {
						spells[3] = append(spells[3], &Spell{
							Game: row[1],
							Name: row[3],
							Rank: row[5],
							Star: int32(star),
							Desc: row[4],
						})
					}
				}
			}
		}
	}
	if len(spells[0]) < spellCounts[0] || len(spells[1]) < spellCounts[1] || len(spells[2])+len(spells[3]) < spellCounts[2] {
		return nil, errors.New("符卡数量不足")
	}
	r := rand.New(rand.NewSource(time.Now().UnixMilli()))
	r.Shuffle(len(spells[0]), func(i, j int) { spells[0][i], spells[0][j] = spells[0][j], spells[0][i] })
	r.Shuffle(len(spells[1]), func(i, j int) { spells[1][i], spells[1][j] = spells[1][j], spells[1][i] })
	spells01 := append(spells[0][:spellCounts[0]:spellCounts[0]], spells[1][:spellCounts[1]]...)
	if len(spells[2]) < spellCounts[2] {
		r.Shuffle(len(spells[3]), func(i, j int) { spells[3][i], spells[3][j] = spells[3][j], spells[3][i] })
		spells[2] = append(spells[2], spells[3][:spellCounts[2]-len(spells[2])]...)
	}
	r.Shuffle(len(spells[2]), func(i, j int) { spells[2][i], spells[2][j] = spells[2][j], spells[2][i] })
	result := make([]*Spell, 25)
	result[0] = spells01[0]
	result[4] = spells01[1]
	result[12] = spells[2][0]
	result[20] = spells[2][1]
	result[24] = spells[2][2]
	r.Shuffle(len(spells01), func(i, j int) { spells01[i], spells01[j] = spells01[j], spells01[i] })
	result[6] = spells01[2]
	result[8] = spells01[3]
	result[16] = spells01[4]
	result[18] = spells01[5]
	spells01 = append(spells01[6:], spells[2][3:spellCounts[2]]...)
	r.Shuffle(len(spells01), func(i, j int) { spells01[i], spells01[j] = spells01[j], spells01[i] })
	j := 0
	for i := range result {
		if result[i] == nil {
			result[i] = spells01[j]
			j++
		}
	}
	return result, nil
}

func (x SpellStatus) isSelectStatus() bool {
	return x == SpellStatus_left_select || x == SpellStatus_right_select || x == SpellStatus_both_select
}

func (x SpellStatus) isLeftStatus() bool {
	return x == SpellStatus_left_select || x == SpellStatus_left_get || x == SpellStatus_both_select || x == SpellStatus_both_get
}

func (x SpellStatus) isRightStatus() bool {
	return x == SpellStatus_right_select || x == SpellStatus_right_get || x == SpellStatus_both_select || x == SpellStatus_both_get
}

func (x SpellStatus) hideLeftSelect() SpellStatus {
	if x == SpellStatus_both_select {
		return SpellStatus_right_select
	}
	return x
}

func (x SpellStatus) hideRightSelect() SpellStatus {
	if x == SpellStatus_both_select {
		return SpellStatus_left_select
	}
	return x
}
