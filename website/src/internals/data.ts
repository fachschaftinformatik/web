export const FORUM_CATEGORIES = ["Ankündigung", "Termin", "Frage", "Diskussion", "Feedback", "Sonstiges"] as const;
export const FORUM_TAGS = ["Hilfe benötigt", "Wichtig", "Klausur", "Organisatorisches"] as const;

export type Member = { id: number; name: string; email: string; img: string | null; };
export type MemberSection = { id: string; title: string; members: Member[]; };

export const MEMBER_SECTIONS: MemberSection[] = [
    { id: "vorstand", title: "Vorstand", members: [
        { id: 1, name: "Eric König (Vorsitz)", email: "vorsitz.informatik@fsv-whs.de", img: null },
        { id: 2, name: "Hannah Willemsen (stellv. Vorsitz)", email: "stellv.vorsitz.informatik@fsv-whs.de", img: null },
        { id: 3, name: "Leon Pearse (Finanzen)", email: "finanzen.informatik@fsv-whs.de", img: null },
    ]},
    { id: "kultur", title: "Kultur", members: [
        { id: 4, name: "Hannah Willemsen", email: "kultur.informatik@fsv-whs.de", img: null },
        { id: 5, name: "Caner Bayram", email: "kultur.informatik@fsv-whs.de", img: null },
        { id: 6, name: "Leon Weihrauch", email: "kultur.informatik@fsv-whs.de", img: null },
        { id: 7, name: "Yannic Leismann", email: "kultur.informatik@fsv-whs.de", img: null },
        { id: 8, name: "Daniel Chudaska", email: "kultur.informatik@fsv-whs.de", img: null },
    ]},
    { id: "infrastruktur", title: "Infrastruktur", members: [{ id: 9, name: "Francesco La Camera", email: "edv.informatik@fsv-whs.de", img: null }]},
    { id: "hochschulpolitik", title: "Hochschulpolitik", members: [{ id: 10, name: "Mike Drost", email: "hopo.informatik@fsv-whs.de", img: null }]},
    { id: "oeffentlichkeit", title: "Öffentlichkeit", members: [
        { id: 11, name: "Linda Gudaqi", email: "oeffentlichkeit.informatik@fsv-whs.de", img: null },
        { id: 12, name: "Hong Nguyen", email: "oeffentlichkeit.informatik@fsv-whs.de", img: null },
    ]},
    { id: "beschaffung", title: "Beschaffung", members: [{ id: 13, name: "Marcel Roith", email: "beschaffung.informatik@fsv-whs.de", img: null }]},
    { id: "kassenwart", title: "Kassenwart", members: [{ id: 15, name: "Eric Seidel", email: "kassenwart.informatik@fsv-whs.de", img: null }]},
    { id: "sport", title: "Sport", members: [
        { id: 16, name: "Caner Bayram", email: "", img: null },
        { id: 17, name: "Yannic Leismann", email: "", img: null },
        { id: 18, name: "Daniel Chudaska", email: "", img: null },
    ]},
];

export const NAV_ITEMS = [
  { label: 'Startseite', href: '/', icon: 'dashboard', isRoute: true },
  { label: 'Diskussionen', href: '/discussions', icon: 'discussions', isRoute: true },
  { label: 'Galerie', href: '/events', icon: 'gallery', isRoute: true },
  { label: 'Archiv', href: '/archive', icon: 'archive', isRoute: true },
  { label: 'Mitglieder', href: '/members', icon: 'team', isRoute: true },
  { label: 'Kontakt', href: '/contact', icon: 'contact', isRoute: true },
];
