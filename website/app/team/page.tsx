'use client';

import Typography from "@mui/material/Typography";
import Card from "@mui/material/Card";
import CardMedia from "@mui/material/CardMedia";
import CardContent from "@mui/material/CardContent";
import Box from "@mui/material/Box";
import Link from "@mui/material/Link";
import AccountCircleIcon from "@mui/icons-material/AccountCircle";
import { Sidebar } from "@components/layout";
import { useAuth } from "@lib/auth";
import { TEAM_SECTIONS, TeamMember } from "@lib/config";

export default function Team() {
  const { user } = useAuth();

  const Section = ({ title, members }: { title: string; members: TeamMember[] }) => {
    const sharedEmail =
      members.length > 1 &&
        members.every((member) => member.email && member.email === members[0]?.email)
        ? members[0]?.email
        : "";

    return (
      <Box sx={{ my: 4 }}>
        <Typography variant="h5" component="h3" gutterBottom>{title}</Typography>
        {sharedEmail && (
          <Link
            href={`mailto:${sharedEmail}`}
            variant="body2"
            underline="hover"
            sx={{
              color: "text.secondary",
              display: "inline-block",
              mb: 2
            }}>
            {sharedEmail}
          </Link>
        )}
        <Box
          sx={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fit, minmax(260px, 320px))",
            justifyContent: "flex-start",
            gap: 4,
          }}
        >
          {members.map((m) => (
            <Card key={m.id} sx={{ borderRadius: 2, overflow: "hidden", height: "100%" }}>

              {m.img ? (
                <CardMedia
                  component="img"
                  height="320"
                  image={m.img}
                  alt={m.name}
                  sx={{ objectFit: "contain", backgroundColor: "grey.100" }}
                />
              ) : (
                <Box
                  sx={{
                    height: 320,
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
                {!sharedEmail && m.email && (
                  <Link
                    href={`mailto:${m.email}`}
                    variant="body2"
                    underline="hover"
                    sx={{
                      color: "text.secondary"
                    }}
                  >
                    {m.email}
                  </Link>
                )}
              </CardContent>
            </Card>
          ))}
        </Box>
      </Box>
    );
  };

  return (
    <Sidebar user={user} title="Unser Team" maxWidth="lg">
      <Box>
        <Box sx={{ mb: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <Box>
            <Typography variant="h4" component="h1" gutterBottom sx={{
              fontWeight: 700
            }}>
              Unser Team
            </Typography>
            <Typography variant="body1" sx={{
              color: "text.secondary"
            }}>
              Mitglieder und ihre Zuständigkeiten
            </Typography>
          </Box>
        </Box>
        {([
          { id: 'elected', title: 'Gewählte Mitglieder' },
          { id: 'project', title: 'Projektstellen (nicht gewählte Mitglieder)' },
        ] as const).map((group) => (
          <Box component="section" key={group.id} aria-labelledby={`team-${group.id}`} sx={{ mt: 5 }}>
            <Typography id={`team-${group.id}`} variant="h4" component="h2">{group.title}</Typography>
            {TEAM_SECTIONS.filter(section => section.group === group.id).map((section) => (
              <Section key={section.id} title={section.title} members={section.members} />
            ))}
          </Box>
        ))}
      </Box>
    </Sidebar>
  );
}
