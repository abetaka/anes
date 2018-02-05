package nespkg

import (
	"fmt"
)

type Mapper interface {
	Init()
	regWrite8(address uint16, val uint8)
}

type MapperMaker func(*Nes) Mapper

var mapperTable = map[int]MapperMaker{
	0: NewMapperBase,
	3: NewMapper003,
}

func MakeMapper(nes *Nes, mapperNum int) (Mapper, error) {
	maker, ok := mapperTable[mapperNum]
	if ok {
		return maker(nes), nil
	} else {
		err := fmt.Errorf("Mapper not supported: mapperNum=%d\n", mapperNum)
		return nil, err
	}
}
