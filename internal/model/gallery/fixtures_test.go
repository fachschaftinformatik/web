package gallery

func Albums() []Album {
	return []Album{
		{ID: "140125680309370881", Title: "LAN Party 2025", OriginalCount: 45, Images: []Image{
			photo("140126343856652289", "Zusammen auf der LAN Party", "Drei Teilnehmende stehen nebeneinander im beleuchteten Veranstaltungsraum.", 1200, 903),
			photo("140126331227602945", "VR ausprobieren", "Eine Person hilft einer anderen beim Aufsetzen einer VR-Brille; beide halten Controller.", 1200, 903),
			photo("140126318313340929", "Am Gaming-PC", "Ein Monitor zeigt das Menü eines Computerspiels; Tastatur, Maus und Kopfhörer liegen auf dem Tisch.", 1200, 903),
			photo("140126304396640257", "Preise auf dem Tisch", "Bedruckte T-Shirts, Tastaturen und Getränkedosen liegen auf einem Tisch bereit.", 1200, 903),
			photo("140126291885031425", "Gemeinsam spielen", "Teilnehmende sitzen an einer langen Tischreihe mit Monitoren und Computern im abendlich beleuchteten Raum.", 1200, 903),
			photo("140126279717355521", "Lichter und Monitore", "Eine Person spielt am Computer neben einem bunt beleuchteten PC-Gehäuse; im Hintergrund stehen weitere Spielstationen.", 1200, 903),
		}},
		{ID: "140117348945231873", Title: "LAN Party 2024", OriginalCount: 133, Images: []Image{
			photo("140123942353371137", "Vor dem Bildschirm", "Drei Teilnehmende stehen mit verschränkten Armen vor einem Monitor.", 1200, 800),
			photo("140123917992853505", "Im Gespräch", "Mehrere Teilnehmende unterhalten sich neben den Computertischen.", 1200, 800),
			photo("140123878637699073", "Am Schachbrett", "Eine Person schaut nachdenklich auf ein Schachbrett auf dem Tisch.", 1200, 800),
			photo("140123825059659777", "Gemeinsam zuschauen", "Mehrere Teilnehmende sitzen vor einer grünen Sitzbank und schauen nach vorne.", 1200, 800),
			photo("140123774497325057", "Zwischen den Spielstationen", "Eine Person mit Brille steht neben einer Reihe von Computertischen; dahinter hängen Lichterketten.", 800, 1200),
			photo("140123726317355009", "Pause am Rand", "Zwei Personen unterhalten sich neben einer grünen Sitzbank und einem Stehtisch.", 1200, 800),
		}},
		{ID: "140115336199405569", Title: "LAN Party 2023", OriginalCount: 50, Images: []Image{
			photo("140116043891736577", "Spielen auf dem Sofa", "Teilnehmende sitzen mit Controllern auf einem Sofa, während andere hinter ihnen zuschauen.", 1200, 628),
			photo("140116033569554433", "Blick aufs Terminal", "Eine Person beugt sich vor einen Monitor mit farbigen Textzeilen in einem Terminal.", 1200, 628),
			photo("140116023507419137", "Gemeinsame Spielpause", "Zwei Teilnehmende sitzen lachend neben einem Computer; weitere Personen stehen dahinter.", 1200, 628),
			photo("140116012379930625", "An den Computertischen", "Teilnehmende sitzen mit Kopfhörern an mehreren Monitoren in einem großen Raum.", 1200, 628),
			photo("140116002028388353", "Mitten im Spiel", "Eine Person bedient Tastatur und Maus vor einem Monitor mit einem Computerspiel.", 1200, 628),
			photo("140115983523119105", "Gruppenfoto", "Fünf Teilnehmende stehen nebeneinander und halten kleine verpackte Gegenstände in den Händen.", 1200, 675),
		}},
	}
}

func Find(id string) (Album, bool) {
	for _, album := range Albums() {
		if album.ID == id {
			return album, true
		}
	}
	return Album{}, false
}

func photo(id, title, description string, width, height int) Image {
	return Image{ID: id, Title: title, Description: description, Width: width, Height: height}
}
