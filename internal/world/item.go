package world

type ItemKind uint8

const (
	ItemHealth ItemKind = iota

	ItemArmor

	ItemBullets

	ItemShells

	ItemRockets

	ItemBackpack

	ItemSoul

	ItemMega

	ItemBerserk

	ItemInvuln

	ItemRadSuit

	ItemKeyRed

	ItemKeyBlue

	ItemKeyYellow
)

func (k ItemKind) IsKey() bool {
	switch k {
	case ItemKeyRed, ItemKeyBlue, ItemKeyYellow:
		return true
	default:
		return false
	}
}

var itemKindNames = map[ItemKind]string{
	ItemHealth:    "health",
	ItemArmor:     "armor",
	ItemBullets:   "bullets",
	ItemShells:    "shells",
	ItemRockets:   "rockets",
	ItemBackpack:  "backpack",
	ItemSoul:      "soulsphere",
	ItemMega:      "megasphere",
	ItemBerserk:   "berserk",
	ItemInvuln:    "invulnerability",
	ItemRadSuit:   "radiation suit",
	ItemKeyRed:    "red key",
	ItemKeyBlue:   "blue key",
	ItemKeyYellow: "yellow key",
}

func (k ItemKind) String() string {
	if s, ok := itemKindNames[k]; ok {
		return s
	}
	return "item"
}

type Item struct {
	Kind ItemKind
	At   Coord
}
