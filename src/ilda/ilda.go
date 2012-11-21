package ilda

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var (
	ErrCorruptInput = errors.New("ilda: corrupt input")

	order = binary.BigEndian
	_d3   = make([]byte, 3)
	_d1   = make([]byte, 1)
)

type Animation struct {
	Frames []*Table
}

func ReadAnimation(r io.Reader) (ani *Animation, err error) {
	frms := make([]*Table, 0, 256)

	for {
		var t *Table
		if t, err = ReadTable(r); err != nil {
			if err == io.EOF {
				break
			}
			return
		}

		if t.Header.Length == 0 {
			break
		}

		frms = append(frms, t)
	}

	return &Animation{frms}, nil
}

func (ani *Animation) String() string {
	return fmt.Sprintf("ILDA animation, %d frames", len(ani.Frames))
}

// Таблица точек (кадр).
type Table struct {
	Header *Header // Заголовок
	Points []Point // Точки
}

// Читает таблицу из потока.
func ReadTable(r io.Reader) (t *Table, err error) {
	var h *Header

	if h, err = ReadHeader(r); err != nil {
		return
	}

	points := make([]Point, h.Length)
	for i := 0; i < int(h.Length); i++ {
		if err = points[i].Read(r); err != nil {
			return
		}
	}

	return &Table{h, points}, nil
}

// Структура заголовка таблицы.
type Header struct {
	Type   Type   // Род содержимого таблицы
	Info   string // Информация
	Length uint16 // Количество элементов в таблице
	Index  uint16 // Порядковый номер таблицы в файле
	Total  uint16 // Всего таблиц или зарезервировано
	Head   byte   // Номер головки сканера
	//_      byte
}

// Читает заголовок из потока.
func ReadHeader(r io.Reader) (h *Header, err error) {
	var v uint32

	if err = binary.Read(r, order, &v); err != nil {
		return
	}
	const MAGIC = 0x494C4441 //0x41444C49
	if v != MAGIC {
		return nil, ErrCorruptInput
	}

	if _, err = r.Read(_d3); err != nil {
		return
	}

	h = new(Header)

	if err = binary.Read(r, order, &h.Type); err != nil {
		return
	}

	b := make([]byte, 16)
	if _, err = r.Read(b); err != nil {
		return
	}
	h.Info = string(bytes.TrimSpace(b))

	if err = binary.Read(r, order, &h.Length); err != nil {
		return
	}
	if err = binary.Read(r, order, &h.Index); err != nil {
		return
	}
	if err = binary.Read(r, order, &h.Total); err != nil {
		return
	}
	if err = binary.Read(r, order, &h.Head); err != nil {
		return
	}

	r.Read(_d1)

	return h, nil
}

func (t *Table) String() string {
	return fmt.Sprintf("frame %d, %s, %d points",
		t.Header.Index, t.Header.Info, t.Header.Length)
}

// Тип таблицы.
type Type byte

const (
	Points3D Type = iota
	Points2D
	Colors
)

type Point struct {
	X      int16
	Y      int16
	Z      int16
	Status Status // Цвет и дополнительная информация
}

func (p *Point) Read(r io.Reader) (err error) {
	if err = binary.Read(r, order, &p.X); err != nil {
		return
	}
	if err = binary.Read(r, order, &p.Y); err != nil {
		return
	}
	if err = binary.Read(r, order, &p.Z); err != nil {
		return
	}
	if err = binary.Read(r, order, &p.Status); err != nil {
		return
	}
	return
}

func (p *Point) String() string {
	b := ' '
	l := ' '
	if p.Status.IsBlank() {
		b = 'b'
	}
	if p.Status.IsLast() {
		l = 'l'
	}

	return fmt.Sprintf("%6d %6d %6d %3d %c %c",
		p.X, p.Y, p.Z, p.Status.GetColor(), b, l)
}

// Цвет точки.
type Color byte

// Статус точки.
type Status uint16

const (
	BLANK_MASK      = 0x4000
	LAST_POINT_MASK = 0x8000
)

// Извлекает цвет точки.
func (s Status) GetColor() Color {
	return Color(byte(s))
}

// Возвращает true, если точка погашена.
func (s Status) IsBlank() bool {
	return (s & BLANK_MASK) == BLANK_MASK
}

// Возвращает true, если точка в таблице последняя.
func (s Status) IsLast() bool {
	return (s & LAST_POINT_MASK) == LAST_POINT_MASK
}
