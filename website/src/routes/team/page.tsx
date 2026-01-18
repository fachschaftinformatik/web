import React from "react";
import {
  Typography,
  Card,
  CardMedia,
  CardContent,
  Box,
  Link,
} from "@mui/material";
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
        <Typography variant="h5" gutterBottom>{title}</Typography>
        {sharedEmail && (
          <Link
            href={`mailto:${sharedEmail}`}
            variant="body2"
            color="text.secondary"
            underline="hover"
            sx={{ display: "inline-block", mb: 2 }}
          >
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
                    color="text.secondary"
                    underline="hover"
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
    <Sidebar user={user} title="Team" maxWidth="lg">
      <Box>
        <Box sx={{ mb: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <Box>
            <Typography variant="h4" component="h1" fontWeight={700} gutterBottom>
              FSV-Team
            </Typography>
            <Typography variant="body1" color="text.secondary">
              Die aktuellen Mitglieder der Fachschaftsvertretung Informatik.
            </Typography>
          </Box>
        </Box>
        {TEAM_SECTIONS.map((section) => (
          <Section key={section.id} title={section.title} members={section.members} />
        ))}
      </Box>
    </Sidebar>
  );
}
