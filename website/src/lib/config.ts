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
    members: TeamMember[];
};

export const TEAM_SECTIONS: TeamSection[] = [
    {
        id: "vorstand",
        title: "Vorstand",
        members: [
            { id: 1, name: "Eric König (Vorsitz)", email: "vorsitz.informatik@fsv-whs.de", img: null },
            { id: 2, name: "Toni Gudaqi (stellv. Vorsitz)", email: "stellv.vorsitz.informatik@fsv-whs.de", img: null },
            { id: 3, name: "Leon Pearse (Finanzen)", email: "finanzen.informatik@fsv-whs.de", img: null },
        ],
    },
    {
        id: "kultur",
        title: "Kultur",
        members: [
            { id: 4, name: "Hannah Willemsen", email: "kultur.informatik@fsv-whs.de", img: null },
            { id: 5, name: "Joel Enrico Sanrio", email: "kultur.informatik@fsv-whs.de", img: null },
            { id: 6, name: "Carlo Niermann", email: "kultur.informatik@fsv-whs.de", img: null },
            { id: 7, name: "Leon Weihrauch", email: "kultur.informatik@fsv-whs.de", img: null },
        ],
    },
    {
        id: "infrastruktur",
        title: "Infrastruktur",
        members: [
            { id: 8, name: "Malte Gaelings", email: "edv.informatik@fsv-whs.de", img: null },
            { id: 9, name: "Birhat Alessandro Moffa", email: "edv.informatik@fsv-whs.de", img: null },
        ],
    },
    {
        id: "hochschulpolitik",
        title: "Hochschulpolitik",
        members: [
            { id: 10, name: "Hong Nguyen", email: "hopo.informatik@fsv-whs.de", img: null },
        ],
    },
    {
        id: "oeffentlichkeit",
        title: "Öffentlichkeit",
        members: [
            { id: 11, name: "Toni Gudaqi", email: "oeffentlichkeit.informatik@fsv-whs.de", img: null },
            { id: 12, name: "Hong Nguyen", email: "oeffentlichkeit.informatik@fsv-whs.de", img: null },
        ],
    },
    {
        id: "beschaffung",
        title: "Beschaffung",
        members: [
            { id: 13, name: "Marcel Roith", email: "beschaffung.informatik@fsv-whs.de", img: null },
            { id: 14, name: "Birhat Alessandro Moffa", email: "beschaffung.informatik@fsv-whs.de", img: null },
        ],
    },
    {
        id: "kassenwart",
        title: "Kassenwart",
        members: [
            { id: 15, name: "Eric Seidel", email: "kassenwart.informatik@fsv-whs.de", img: null },
        ],
    },
    {
        id: "sport",
        title: "Sport",
        members: [
            { id: 16, name: "Zafer Elemen", email: "", img: null },
            { id: 17, name: "Caner Bayram", email: "", img: null },
            { id: 18, name: "Imad Boubkar", email: "", img: null },
            { id: 19, name: "Sami Boubkar", email: "", img: null },
        ],
    },
];
