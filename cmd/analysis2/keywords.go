package main

import "sort"
import "strings"
import "sync"

// this finds (usually short) words which may exist in other words but other words don't exist in them
func Keywords(threading int, dict [][2]string) (ret [][2]string) {
	sort.Slice(dict, func(i, j int) bool {
		return len(dict[i][0])+len(dict[i][1]) < len(dict[j][0])+len(dict[j][1])
	})
	var mut sync.Mutex
	loop(dict, threading, func(word1, word2 string) {
		for i := range dict {
			if len(dict[i][0]) < len(word1) {
				if strings.Contains(word1, dict[i][0]) {
					return
				}
			}
			if len(dict[i][1]) < len(word2) {
				if strings.Contains(word2, dict[i][1]) {
					return
				}
			}
			if len(dict[i][0])+len(dict[i][1]) > len(word1)+len(word2) {
				break
			}
		}
		mut.Lock()
		ret = append(ret, [2]string{word1, word2})
		mut.Unlock()
	})
	sort.Slice(ret, func(i, j int) bool {
		return len(ret[i][0])+len(ret[i][1]) < len(ret[j][0])+len(ret[j][1])
	})
	return ret
}
