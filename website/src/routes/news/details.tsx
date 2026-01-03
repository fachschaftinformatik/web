import { useParams, useNavigate } from "react-router-dom";
import { Box, Typography, Link, Button, Chip, Stack } from "@mui/material";
import { useMemo } from "react";
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import { Sidebar } from "@components/layout";
import { useAuth } from "@lib/auth";
import { NEWS_DATA } from "@lib/data";

export default function NewsDetail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { user } = useAuth();

  const news = useMemo(() => {
    const custom = JSON.parse(localStorage.getItem("custom-news") || "[]");
    const allNews = [...NEWS_DATA, ...custom];
    return allNews.find((item) => item.id === Number(id)) || null;
  }, [id]);


  if (!news) {
    return (
      <Sidebar user={user} title="Beitrag nicht gefunden">
        <Box sx={{ pt: 10 }}>
          <Typography variant="h5">Beitrag nicht gefunden.</Typography>
          <Button startIcon={<ArrowBackIcon />} onClick={() => navigate("/news")} sx={{ mt: 2 }}>
            Zurück zur Übersicht
          </Button>
        </Box>
      </Sidebar>
    );
  }

  return (
    <Sidebar user={user} title="News Detail">
      <Button startIcon={<ArrowBackIcon />} onClick={() => navigate("/news")} sx={{ mb: 3 }}>
        Zurück
      </Button>

      <Typography variant="h3" sx={{ mb: 2, fontWeight: 700 }}>
        {news.title}
      </Typography>

      <Stack direction="row" alignItems="center" spacing={2} sx={{ mb: 3 }}>
        <Typography variant="subtitle1" color="text.secondary">
          {news.date}
        </Typography>
        {news.tags?.map((tag: string) => (
          <Chip key={tag} label={tag} size="small" variant="outlined" />
        ))}
      </Stack>

      {news.image && (
        <Box
          component="img"
          src={news.image}
          alt={news.title}
          sx={{
            width: "100%",
            maxHeight: 500,
            objectFit: "cover",
            borderRadius: 3,
            boxShadow: 2,
            mb: 4
          }}
        />
      )}

      <Typography variant="h6" sx={{ mb: 2, fontStyle: 'italic', color: 'text.secondary' }}>
        {news.summary}
      </Typography>

      <Typography variant="body1" sx={{ mb: 4, whiteSpace: "pre-line", fontSize: '1.1rem', lineHeight: 1.7 }}>
        {news.content}
      </Typography>

      {news.links && news.links.length > 0 && (
        <Box sx={{ mt: 4, p: 3, bgcolor: 'background.paper', borderRadius: 2, border: '1px solid', borderColor: 'divider' }}>
          <Typography variant="h6" sx={{ mb: 2 }}>
            Weiterführende Links
          </Typography>
          <Stack spacing={1}>
            {news.links.map((link: string, index: number) => (
              <Link key={index} href={link} target="_blank" rel="noopener" underline="hover" color="primary">
                {link}
              </Link>
            ))}
          </Stack>
        </Box>
      )}

      {news.pdf && (
        <Box sx={{ mt: 4 }}>
          <Button
            variant="outlined"
            color="primary"
            href={news.pdf}
            target="_blank"
            sx={{ py: 1.5, px: 3 }}
          >
            📄 {news.pdfName || "PDF Anhang öffnen"}
          </Button>
        </Box>
      )}
    </Sidebar>
  );
}
