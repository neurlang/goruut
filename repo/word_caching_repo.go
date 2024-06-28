package repo

import (
	"crypto/rand"
	"encoding/binary"
	"github.com/maypok86/otter"
	"github.com/neurlang/goruut/helpers/log"
	"github.com/spaolacci/murmur3"
	"time"
)
import . "github.com/martinarisk/di/dependency_injection"

type IWordCachingRepository interface {
	HashWord(lang, word string) uint64
	StoreWord(value map[uint64]string, hash uint64)
	LoadWord(hash uint64) (word map[uint64]string)
}
type WordCachingRepository struct {
	seed  uint32
	cache otter.Cache[uint64, string]
}

func (r WordCachingRepository) LoadWord(hash uint64) (word map[uint64]string) {
	value, _ := r.cache.Get(hash)
	if value == "" {
		return nil
	}
	word = make(map[uint64]string)
	length := binary.LittleEndian.Uint32([]byte(value[0:4]))
	end := 4 + length*12
	for i := uint32(0); i < length; i++ {
		k := binary.LittleEndian.Uint64([]byte(value[3*i+4 : 3*i+12]))
		l := binary.LittleEndian.Uint32([]byte(value[3*i+12 : 3*i+16]))
		word[k] = value[end : end+l]
		end += l
	}
	return word
}
func (r WordCachingRepository) StoreWord(value map[uint64]string, hash uint64) {

	var buf, data []byte
	var num4 [4]byte
	var num8 [8]byte
	binary.LittleEndian.PutUint32(num4[:], uint32(len(value)))
	buf = append(buf, num4[:]...)

	for k, v := range value {
		binary.LittleEndian.PutUint64(num8[:], uint64(k))
		buf = append(buf, num8[:]...)
		binary.LittleEndian.PutUint32(num4[:], uint32(len(v)))
		buf = append(buf, num4[:]...)
		data = append(data, []byte(v)...)
	}

	val := string(buf) + string(data)

	r.cache.Set(hash, val)
}
func (r WordCachingRepository) HashWord(lang, word string) uint64 {

	str := word + "\x00" + lang

	return murmur3.Sum64WithSeed([]byte(str), r.seed)
}

func NewWordCachingRepository(di *DependencyInjection) *WordCachingRepository {

	var buf [4]byte
	rand.Read(buf[:])
	seed := binary.LittleEndian.Uint32(buf[:])

	// create a cache with capacity equal to 10000 elements
	cache := log.Error1(otter.MustBuilder[uint64, string](10_000).
		CollectStats().
		Cost(func(key uint64, value string) uint32 {
			return 1
		}).
		WithTTL(time.Hour).
		Build())

	return &WordCachingRepository{
		seed:  seed,
		cache: cache,
	}
}

var _ IWordCachingRepository = &WordCachingRepository{}
