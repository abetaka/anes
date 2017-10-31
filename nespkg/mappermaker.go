package nespkg

type Mapper interface {
	Init()
	regWrite8(address uint16, val uint8)
}

type MapperMaker func(*Nes) Mapper

var mapperTable = map[int]MapperMaker{
	0: NewMapperBase,
	3: NewMapper003,
}

func MakeMapper(nes *Nes, mapperNum int) Mapper {
	return mapperTable[mapperNum](nes)
}
