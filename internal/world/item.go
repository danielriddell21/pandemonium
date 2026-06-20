package world

// ItemKind enumerates the collectibles a level can hold: consumables that top up
// the player's reserves and coloured keycards that open their matching doors.
type ItemKind uint8

const (
	// ItemHealth restores health when collected.
	ItemHealth ItemKind = iota
	// ItemArmor adds armour, which later absorbs a share of incoming damage.
	ItemArmor
	// ItemBullets replenishes pistol ammunition.
	ItemBullets
	// ItemShells replenishes shotgun ammunition.
	ItemShells
	// ItemRockets replenishes rocket-launcher ammunition.
	ItemRockets
	// ItemBackpack doubles ammo capacity and tops the player up.
	ItemBackpack
	// ItemKeyRed opens red-locked doors.
	ItemKeyRed
	// ItemKeyBlue opens blue-locked doors.
	ItemKeyBlue
	// ItemKeyYellow opens yellow-locked doors.
	ItemKeyYellow
)

// IsKey reports whether the kind is a keycard rather than a consumable.
func (k ItemKind) IsKey() bool {
	switch k {
	case ItemKeyRed, ItemKeyBlue, ItemKeyYellow:
		return true
	default:
		return false
	}
}

// String returns a stable label for the kind, suitable for logs and HUD text.
func (k ItemKind) String() string {
	switch k {
	case ItemHealth:
		return "health"
	case ItemArmor:
		return "armor"
	case ItemBullets:
		return "bullets"
	case ItemShells:
		return "shells"
	case ItemRockets:
		return "rockets"
	case ItemBackpack:
		return "backpack"
	case ItemKeyRed:
		return "red key"
	case ItemKeyBlue:
		return "blue key"
	case ItemKeyYellow:
		return "yellow key"
	default:
		return "item"
	}
}

// Item is a collectible placed on a walkable cell of a level.
type Item struct {
	Kind ItemKind
	At   Coord
}
