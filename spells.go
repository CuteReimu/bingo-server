package main

import (
	"github.com/Touhou-Freshman-Camp/bingo-server/arrays"
	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func RandSpells(games []string) ([]*Spell, error) {
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
					inGame := arrays.Contains(games, strings.TrimSpace(row[1]))
					if star > 0 && star <= 3 && inGame {
						spells[star-1] = append(spells[star-1], &Spell{
							Game: row[1],
							Name: row[3],
							Rank: row[5],
							Star: int32(star),
						})
					}
					if star == 3 && !inGame {
						spells[3] = append(spells[3], &Spell{
							Game: row[1],
							Name: row[3],
							Rank: row[5],
							Star: int32(star),
						})
					}
				}
			}
		}
	}
	if len(spells[0]) < 10 || len(spells[1]) < 10 || len(spells[2])+len(spells[3]) < 5 {
		return nil, errors.New("符卡数量不足")
	}
	r := rand.New(rand.NewSource(time.Now().UnixMilli()))
	arrays.ShuffleN(r, spells[0], 10)
	arrays.ShuffleN(r, spells[1], 10)
	spells01 := append(spells[0][:10:10], spells[1][:10]...)
	arrays.ShuffleN(r, spells01, len(spells01))
	if len(spells[2]) < 5 {
		arrays.ShuffleN(r, spells[3], 5-len(spells[2]))
		spells[2] = append(spells[2], spells[3][:5-len(spells[2])]...)
	}
	arrays.ShuffleN(r, spells[2], 5)
	idx := []int{0, 1, 3, 4}
	arrays.ShuffleN(r, idx, len(idx))
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
