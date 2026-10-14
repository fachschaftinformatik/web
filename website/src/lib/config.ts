export const FORUM_CATEGORIES = [
    "Ankündigung",
    "Termin",
    "Frage",
    "Diskussion",
    "Feedback",
    "Sonstiges"
] as const;

export const FORUM_TAGS = [
    "Hilfe benötigt",
    "Wichtig",
    "Klausur",
    "Organisatorisches"
] as const;

export type TeamMember = {
    id: number;
    name: string;
    email: string;
    img: string | null;
};

export type TeamSection = {
    id: string;
    title: string;
    group: "elected" | "project";
    members: TeamMember[];
};

export const TEAM_SECTIONS: TeamSection[] = [
    {
        id: "vorstand",
        title: "Vorstand",
        group: "elected",
        members: [
            { id: 1, name: "Eric Nik König (Vorsitz)", email: "vorsitz.informatik@fsv-whs.de", img: null },
            { id: 2, name: "Linda Gudaqi (Stellvertretender Vorsitz)", email: "stellv.vorsitz.informatik@fsv-whs.de", img: null },
            { id: 3, name: "Leon Pearse (Finanzen)", email: "finanzen.informatik@fsv-whs.de", img: null },
        ],
    },
    {
        id: "kultur",
        title: "Kultur",
        group: "elected",
        members: [
            { id: 4, name: "Hannah Willemsen", email: "kultur.informatik@fsv-whs.de", img: null },
            { id: 5, name: "Joel Enrico Soriano", email: "kultur.informatik@fsv-whs.de", img: null },
            { id: 6, name: "Carlo Niermann", email: "kultur.informatik@fsv-whs.de", img: null },
            { id: 7, name: "Leon Weihrauch", email: "kultur.informatik@fsv-whs.de", img: null },
        ],
    },
    {
        id: "infrastruktur",
        title: "Infrastruktur",
        group: "elected",
        members: [
            { id: 8, name: "Malte Gaelings", email: "edv.informatik@fsv-whs.de", img: null },
            { id: 9, name: "Birhat Alessandro Moffa", email: "edv.informatik@fsv-whs.de", img: null },
            { id: 16, name: "Zafer Elemen", email: "edv.informatik@fsv-whs.de", img: null },
        ],
    },
    {
        id: "hochschulpolitik",
        title: "Hochschulpolitik",
        group: "elected",
        members: [
            { id: 10, name: "Thuy Hong Nguyen", email: "hopo.informatik@fsv-whs.de", img: null },
        ],
    },
    {
        id: "oeffentlichkeit",
        title: "Öffentlichkeit",
        group: "elected",
        members: [
            { id: 11, name: "Linda Gudaqi", email: "oeffentlichkeit.informatik@fsv-whs.de", img: null },
            { id: 12, name: "Thuy Hong Nguyen", email: "oeffentlichkeit.informatik@fsv-whs.de", img: null },
        ],
    },
    {
        id: "beschaffung",
        title: "Beschaffung",
        group: "elected",
        members: [
            { id: 20, name: "Leon Weihrauch", email: "beschaffung.informatik@fsv-whs.de", img: null },
            { id: 13, name: "Marcel Roith", email: "beschaffung.informatik@fsv-whs.de", img: null },
            { id: 14, name: "Birhat Alessandro Moffa", email: "beschaffung.informatik@fsv-whs.de", img: null },
        ],
    },
    {
        id: "sport",
        title: "Sport",
        group: "elected",
        members: [
            { id: 17, name: "Caner Bayram", email: "", img: null },
            { id: 18, name: "Imad Boubkar", email: "", img: null },
            { id: 19, name: "Sami Boubkar", email: "", img: null },
        ],
    },
    {
        id: "projekt-fotografie",
        title: "Fotografie",
        group: "project",
        members: [
            { id: 21, name: "Kubilay Cakmakkaya", email: "", img: null },
        ],
    },
    {
        id: "projekt-infrastruktur-webseite",
        title: "Infrastruktur & Webseite",
        group: "project",
        members: [
            { id: 22, name: "Francesco La Camera", email: "", img: null },
        ],
    },
    {
        id: "projekt-sport",
        title: "Sport",
        group: "project",
        members: [
            { id: 23, name: "Mohamed Boubkar El Hachni", email: "", img: null },
        ],
    },
    {
        id: "projekt-licht-ton",
        title: "Licht & Ton",
        group: "project",
        members: [
            { id: 24, name: "Yannic Leismann", email: "", img: null },
        ],
    },
];
