import { useParams } from "react-router-dom";
import { newsDaten } from "./newsLayout";
import { Box, Typography, Container, Link } from "@mui/material";

export default function NewsDetail() {
  const { id } = useParams();
  const news = newsDaten.find((item) => item.id === Number(id));

  if (!news) {
    return <Typography variant="h5" sx={{ p: 4 }}>News nicht gefunden.</Typography>;
  }

  return (
    <Container sx={{ pt: 10, pb: 10 }}>
      {/* Titel */}
      <Typography variant="h3" sx={{ mb: 2 }}>
        {news.title}
      </Typography>

      {/* Datum */}
      <Typography variant="body2" sx={{ color: "text.secondary", mb: 4 }}>
        {news.date}
      </Typography>

      {/* Bild in voller Größe */}
      <Box
        component="img"
        src={news.image}
        alt={news.title}
        sx={{ width: "100%", maxHeight: 500, objectFit: "cover", borderRadius: 2, mb: 4 }}
      />

      {/* Inhalt */}
      <Typography variant="body1" sx={{ mb: 4, whiteSpace: "pre-line" }}>
        {news.content}
      </Typography>

      {/* Optional: Links */}
      {news.links && news.links.length > 0 && (
        <Box sx={{ mt: 4 }}>
          <Typography variant="h6" sx={{ mb: 1 }}>Links:</Typography>
          {news.links.map((link, index) => (
            <Typography key={index}>
              <Link href={link} target="_blank" rel="noopener">
                {link}
              </Link>
            </Typography>
          ))}
        </Box>
      )}
    </Container>
  );
}