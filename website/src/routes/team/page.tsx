import React from "react";
import { Container, Typography, Card, CardMedia, CardContent, Box, Link } from "@mui/material";
import AccountCircleIcon from '@mui/icons-material/AccountCircle';
import { Sidebar } from "@components/layout";
import { useAuth } from "@lib/auth";

export default function Team() {
  const { user } = useAuth();
  const vorstand = [
    { id: 1, name: "Vorname Nachname", email: "mail1@fsv.de", img: null },
    { id: 2, name: "Vorname Nachname", email: "mail2@fsv.de", img: null },
    { id: 3, name: "Vorname Nachname", email: "mail3@fsv.de", img: null },
  ];

  const edv = [
    { id: 4, name: "Vorname Nachname", email: "mail4@fsv.de", img: null },
    { id: 5, name: "Vorname Nachname", email: "mail5@fsv.de", img: null },
    { id: 6, name: "Vorname Nachname", email: "mail6@fsv.de", img: null },
  ];

  const hopo = [
    { id: 7, name: "Vorname Nachname", email: "mail7@fsv.de", img: null },
    { id: 8, name: "Vorname Nachname", email: "mail8@fsv.de", img: null },
    { id: 9, name: "Vorname Nachname", email: "mail9@fsv.de", img: null },
  ];

  // --- HIER IST DIE KORREKTUR ---
  // Alle folgenden Arrays sind jetzt auch mit Platzhaltern gefüllt.

  const kassenwart = [
    { id: 10, name: "Vorname Nachname", email: "mail10@fsv.de", img: null },
    { id: 11, name: "Vorname Nachname", email: "mail11@fsv.de", img: null },
    { id: 12, name: "Vorname Nachname", email: "mail12@fsv.de", img: null },
  ];

  const kultur = [
    { id: 13, name: "Vorname Nachname", email: "mail13@fsv.de", img: null },
    { id: 14, name: "Vorname Nachname", email: "mail14@fsv.de", img: null },
    { id: 15, name: "Vorname Nachname", email: "mail15@fsv.de", img: null },
  ];

  const oeffentlichkeit = [
    { id: 16, name: "Vorname Nachname", email: "mail16@fsv.de", img: null },
    { id: 17, name: "Vorname Nachname", email: "mail17@fsv.de", img: null },
    { id: 18, name: "Vorname Nachname", email: "mail18@fsv.de", img: null },
  ];
  
  const soziales = [
    { id: 19, name: "Vorname Nachname", email: "mail19@fsv.de", img: null },
    { id: 20, name: "Vorname Nachname", email: "mail20@fsv.de", img: null },
    { id: 21, name: "Vorname Nachname", email: "mail21@fsv.de", img: null },
  ];

  const sport = [
    { id: 22, name: "Vorname Nachname", email: "mail22@fsv.de", img: null },
    { id: 23, name: "Vorname Nachname", email: "mail23@fsv.de", img: null },
    { id: 24, name: "Vorname Nachname", email: "mail24@fsv.de", img: null },
  ];


  // --- ANGEPASSTE SECTION-KOMPONENTE (unverändert) ---
  const Section = ({ title, members }: { title: string; members: typeof vorstand }) => (
    <Box sx={{ my: 4 }}>
      <Typography variant="h5" gutterBottom>{title}</Typography>
      <Box
        sx={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fit, minmax(260px, 1fr))",
          gap: 4, 
        }}
      >
        {members.map((m) => (
          <Card key={m.id} sx={{ borderRadius: 2, overflow: "hidden", height: "100%" }}>
            
            {m.img ? (
              <CardMedia 
                component="img" 
                height="280" 
                image={m.img} 
                alt={m.name} 
                sx={{ objectFit: "cover" }} 
              />
            ) : (
              <Box 
                sx={{
                  height: 280,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  backgroundColor: 'grey.200',
                }}
              >
                <AccountCircleIcon 
                  sx={{ 
                    fontSize: 200, 
                    color: 'grey.400' 
                  }} 
                />
              </Box>
            )}
            <CardContent>
              <Typography variant="subtitle1">{m.name}</Typography>
              <Link
                href={`mailto:${m.email}`} 
                variant="body2"
                color="text.secondary" 
                underline="hover" 
              >
                {m.email} 
              </Link>
            </CardContent>
          </Card>
        ))}
      </Box>
    </Box>
  );
  // --- RENDER-TEIL (unverändert) ---
  return (
    <Sidebar user={user} title="Team">
    <Container sx={{ py: 4 }}>
      <Typography variant="h4" align="center" gutterBottom>FSV-Team</Typography>
      <Section title="Vorstand" members={vorstand} />
      <Section title="EDV" members={edv} />
      <Section title="Hopo" members={hopo} />
      <Section title="Kassenwart" members={kassenwart} />
      <Section title="Kultur" members={kultur} />
      <Section title="Öffentlichkeit" members={oeffentlichkeit} />
      <Section title="Soziales" members={soziales} />
      <Section title="Sport" members={sport} />
    </Container>
    </Sidebar>
  );
}
