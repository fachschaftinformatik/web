import Typography from "@mui/material/Typography";
import Card from "@mui/material/Card";
import CardMedia from "@mui/material/CardMedia";
import CardContent from "@mui/material/CardContent";
import Box from "@mui/material/Box";
import Link from "@mui/material/Link";
import AccountCircleIcon from "@mui/icons-material/AccountCircle";
import Page from "@components/Page";
import { MEMBER_SECTIONS, Member } from "@internals/data";

export default function Members() {
  const Section = ({ title, members }: { title: string; members: Member[] }) => {
    const shared = members.length > 1 && members.every(m => m.email && m.email === members[0]?.email) ? members[0]?.email : "";

    return (
      <Box sx={{ my: 4 }}>
        <Typography variant="h5" gutterBottom>{title}</Typography>
        {shared && <Link href={`mailto:${shared}`} variant="body2" color="text.secondary" underline="hover" sx={{ display: "inline-block", mb: 2 }}>{shared}</Link>}
        <Box sx={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(260px, 320px))", gap: 4 }}>
          {members.map(m => (
            <Card key={m.id} sx={{ borderRadius: 2, height: "100%" }}>
              {m.img ? <CardMedia component="img" height="320" image={m.img} alt={m.name} sx={{ objectFit: "contain", bgcolor: "grey.100" }} /> :
                <Box sx={{ height: 320, display: 'flex', alignItems: 'center', justifyContent: 'center', bgcolor: 'grey.200' }}>
                  <AccountCircleIcon sx={{ fontSize: 200, color: 'grey.400' }} />
                </Box>}
              <CardContent>
                <Typography variant="subtitle1">{m.name}</Typography>
                {!shared && m.email && <Link href={`mailto:${m.email}`} variant="body2" color="text.secondary" underline="hover">{m.email}</Link>}
              </CardContent>
            </Card>
          ))}
        </Box>
      </Box>
    );
  };

  return (
    <Page title="Mitglieder" description="Die aktuellen Mitglieder der Fachschaftsvertretung Informatik.">
      {MEMBER_SECTIONS.map(s => <Section key={s.id} title={s.title} members={s.members} />)}
    </Page>
  );
}
