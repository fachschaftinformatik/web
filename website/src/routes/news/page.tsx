import { useState, useMemo } from 'react';
import {
  Card, Button, CardMedia, CardContent, CardActions, Avatar, IconButton, Typography,
  Container, List, ListItem, Divider, ListItemText, ListItemAvatar,
  Box, Chip, Stack, Paper, InputBase, CardActionArea
} from '@mui/material';
import Grid from '@mui/material/Grid';
import FavoriteIcon from '@mui/icons-material/Favorite';
import StarBorderIcon from '@mui/icons-material/StarBorder';
import SearchIcon from '@mui/icons-material/Search';
import AddIcon from "@mui/icons-material/Add";
import DeleteIcon from "@mui/icons-material/Delete";
import { useNavigate, Link, Outlet } from 'react-router-dom';
import { Sidebar } from '@components/layout';
import { useAuth } from '@lib/auth';

// --- DATEN (Exportiert für details.tsx) ---
export type NewsItem = {
  id: number;
  title: string;
  date: string;
  createdAt?: string;
  image: string;
  summary: string;
  content: string;
  links: string[];
  isNew?: boolean;
  tags: string[];
  pdf?: string | null;
  pdfName?: string | null;
  isNewAt?: number;
};

export const newsDaten: NewsItem[] = [
  {
    id: 1,
    title: "6-Year-Old Horse Dies at Belmont Park After Race Injury",
    date: "September 14, 2016",
    image: "https://images.unsplash.com/photo-1607746882042-944635dfe10e?auto=format&fit=crop&w=400&h=200&q=80",
    summary: "A tragic accident occurred during the race, raising concerns about animal safety...",
    content: "Lorem ipsum dolor sit amet...",
    links: ["https://conference2025.com", "https://more-info.com"],
    tags: ["Events", "Jobs"],
    isNewAt: Date.now() - 1 * 24 * 60 * 60 * 1000
  },
  {
    id: 2,
    title: "Tech Conference 2025 Kicks Off in Berlin",
    date: "September 12, 2025",
    image: "https://images.unsplash.com/photo-1607746882042-944635dfe10e?auto=format&fit=crop&w=400&h=200&q=80",
    summary: "Industry leaders gather to discuss the future of AI, robotics, and quantum computing...",
    content: "",
    links: ["https://conference2025.com"],
    tags: ["Prüfungen", "Studium"],
    isNewAt: Date.now() - 1 * 24 * 60 * 60 * 1000
  },
  {
    id: 3,
    title: "New Climate Agreement Signed",
    date: "September 10, 2025",
    image: "https://images.unsplash.com/photo-1607746882042-944635dfe10e?auto=format&fit=crop&w=400&h=200&q=80",
    summary: "World leaders agreed on a historic pact aiming to reduce emissions worldwide...",
    content: "",
    links: [],
    tags: ["Events", "Studium"],
    isNewAt: Date.now() - 20 * 24 * 60 * 60 * 1000
  },
  {
    id: 4,
    title: "-Year-Old Horse Dies at Belmont Park After Race Injury",
    date: "September 8, 2025",
    image: "https://images.unsplash.com/photo-1607746882042-944635dfe10e?auto=format&fit=crop&w=400&h=200&q=80",
    summary: "Scientists reported a promising new therapy with encouraging early trial results...",
    content: "",
    links: [],
    tags: ["Studium"],
    isNewAt: Date.now() - 20 * 24 * 60 * 60 * 1000
  },
  {
    id: 5,
    title: "Breakthrough in Cancer Research",
    date: "September 8, 2025",
    image: "https://images.unsplash.com/photo-1607746882042-944635dfe10e?auto=format&fit=crop&w=400&h=200&q=80",
    summary: "Scientists reported a promising new therapy with encouraging early trial results...",
    content: "",
    links: [],
    tags: ["Events", "Studium"],
    isNewAt: Date.now() - 20 * 24 * 60 * 60 * 1000
  },
];

// --- COMPONENTS ---
function CustomizedInputBase() {
  return (
    <Paper
      component="form"
      sx={{ p: '2px 4px', display: 'flex', alignItems: 'center', width: "100%", borderRadius: 2 }}
    >
      <IconButton type="button" sx={{ p: '10px' }} aria-label="search">
        <SearchIcon />
      </IconButton>
      <InputBase
        sx={{ ml: 1, flex: 1 }}
        placeholder="Suche"
        inputProps={{ 'aria-label': 'search' }}
      />
      <Divider sx={{ height: 28, m: 0.5 }} orientation="vertical" />
    </Paper>
  );
}

function ClickableChips({ selectedTag, setSelectedTag }: { selectedTag: string, setSelectedTag: (tag: string) => void }) {
  const tags = ["Alle", "Events", "Studium", "Prüfungen", "Jobs"];
  return (
    <Stack direction="row" spacing={1}>
      {tags.map((tag) => (
        <Chip
          key={tag}
          label={tag}
          onClick={() => setSelectedTag(tag)}
          color={selectedTag === tag ? "success" : "default"}
          variant={selectedTag === tag ? "filled" : "outlined"}
        />
      ))}
    </Stack>
  );
}

type CardProps = NewsItem & {
  isLiked: boolean;
  onToggleLike: (id: number) => void;
  onDelete: (id: number) => void;
  isAdmin: boolean;
};

function ReviewCard({ id, title, date, image, summary, isLiked, tags, isNew, isAdmin, onToggleLike, onDelete }: CardProps) {
  const navigate = useNavigate();
  return (
    <Card sx={{ width: '100%', height: '100%', display: 'flex', flexDirection: 'column', borderRadius: 3 }}>
      <CardActionArea sx={{ flexGrow: 1 }} onClick={() => navigate(`/news/${id}`)}>
        {image && (
          <CardMedia
            component="img"
            height="200"
            image={image}
            alt={title}
          />
        )}
        <CardContent sx={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
          <Box sx={{ mb: 1 }}>
            <Typography variant="h6"
              sx={{
                display: '-webkit-box',
                WebkitLineClamp: 2,
                WebkitBoxOrient: 'vertical',
                overflow: 'hidden',
                lineHeight: '1.2rem',
                minHeight: '2.4rem'
              }}
            >
              {title}
            </Typography>
            <Typography variant="caption" sx={{ color: 'text.secondary', display: 'block', mt: 1 }}>
              {date}
            </Typography>
          </Box>
          <Box>
            <Stack direction="row" spacing={1} mb={1}>
              {tags && tags.length > 0 ? (
                tags.map((tag) => <Chip key={tag} label={tag} size="small" />)
              ) : (
                <Chip label="Allgemein" size="small" variant="filled" />
              )}
            </Stack>
            <Typography
              variant="body2"
              sx={{
                color: 'text.secondary',
                display: '-webkit-box',
                WebkitLineClamp: 3,
                WebkitBoxOrient: 'vertical',
                overflow: 'hidden',
              }}
            >
              {summary}
            </Typography>
          </Box>
        </CardContent>
      </CardActionArea>
      <CardActions disableSpacing sx={{ justifyContent: "space-between" }}>
        <Box>
          <IconButton onClick={(e) => { e.stopPropagation(); onToggleLike(id); }} aria-label="add to favorites">
            {isLiked ? <FavoriteIcon color="error" /> : <FavoriteIcon />}
          </IconButton>
          {isNew && (
            <Chip
              icon={<StarBorderIcon />}
              label="Neu"
              color="warning"
              size="small"
              sx={{ ml: 1 }}
            />
          )}
        </Box>

        {isAdmin && (
          <Button
            variant="text"
            color="error"
            endIcon={<DeleteIcon />}
            onClick={(e) => { e.stopPropagation(); onDelete(id); }}
            sx={{ fontSize: "0.75rem" }}
          >
            Löschen
          </Button>
        )}

      </CardActions>
    </Card>
  );
}

type FavoriteListProps = {
  likedIds: number[];
  allNews: NewsItem[];
};

function FavoriteList({ likedIds, allNews }: FavoriteListProps) {
  const likedItems = allNews.filter((item) => likedIds.includes(item.id));
  return (
    <List sx={{ width: '100%' }}>
      {likedItems.map((item) => (
        <ListItem key={item.id}>
          {item.image && (
            <ListItemAvatar>
              <Avatar src={item.image} variant="square" sx={{ width: 60, height: 60, mr: 2, borderRadius: 1 }} />
            </ListItemAvatar>
          )}
          <ListItemText
            primary={
              <Link
                to={`/news/${item.id}`}
                style={{ textDecoration: 'none', color: 'inherit' }}
              >
                <Typography variant="body2" sx={{
                  fontWeight: 'bold',
                  '&:hover': { textDecoration: 'underline' }
                }}>
                  {item.title}
                </Typography>
              </Link>
            }
            secondary={item.date}
          />
        </ListItem>
      ))}
      {likedItems.length === 0 && (
        <Typography variant="body2" color="text.secondary" sx={{ p: 2 }}>
          Noch keine Favoriten.
        </Typography>
      )}
    </List>
  );
}

export default function NewsLayout() {
  const { user } = useAuth();
  const isAdmin = user?.role === "admin";
  const [selectedTag, setSelectedTag] = useState<string>("Alle");

  // Lokale Daten für Likes und eigene Beiträge (um Re-Renders zu triggern)
  const [localLikes, setLocalLikes] = useState<number[]>(() => {
    const saved = localStorage.getItem('likedNews');
    return saved ? JSON.parse(saved) : [];
  });

  const [localCustomNews, setLocalCustomNews] = useState<NewsItem[]>(() => {
    return JSON.parse(localStorage.getItem("custom-news") || "[]");
  });

  const [now] = useState(() => Date.now());

  const allNews = useMemo(() => {
    const merged = [...newsDaten, ...localCustomNews];

    const updatedNews = merged.map(item => {
      const created = item.isNewAt ?? (typeof item.createdAt === 'string' ? new Date(item.createdAt).getTime() : now);
      const ageInDays = (now - created) / (1000 * 60 * 60 * 24);
      return {
        ...item,
        isNew: ageInDays <= 7,
      };
    });

    updatedNews.sort((a, b) => {
      const timeA = typeof a.isNewAt === 'number' ? a.isNewAt : (typeof a.createdAt === 'string' ? new Date(a.createdAt).getTime() : new Date(a.date).getTime());
      const timeB = typeof b.isNewAt === 'number' ? b.isNewAt : (typeof b.createdAt === 'string' ? new Date(b.createdAt).getTime() : new Date(b.date).getTime());
      return timeB - timeA;
    });

    return updatedNews;
  }, [localCustomNews, now]);

  const handleToggleLike = (id: number) => {
    setLocalLikes((prev) => {
      const updatedLikes = prev.includes(id)
        ? prev.filter((itemId) => itemId !== id)
        : [...prev, id];
      localStorage.setItem('likedNews', JSON.stringify(updatedLikes));
      return updatedLikes;
    });
  };

  const handleDelete = (id: number) => {
    const stored = JSON.parse(localStorage.getItem("custom-news") || "[]");
    const filteredStored = stored.filter((item: NewsItem) => item.id !== id);
    localStorage.setItem("custom-news", JSON.stringify(filteredStored));
    setLocalCustomNews(filteredStored);
  };

  const filteredNews = selectedTag === "Alle"
    ? allNews
    : allNews.filter((n) => n.tags?.includes(selectedTag));

  return (
    <Sidebar user={user} title="Beiträge">
      <Box sx={{ display: 'flex', flexDirection: { xs: 'column', lg: 'row' }, px: 2 }}>
        {/* Hauptinhalt */}
        <Box sx={{ flex: 1, maxWidth: "1200px", mr: { lg: 4 }, mb: 4 }}>
          <Container sx={{ pt: 4, pb: 2 }}>
            <Box
              sx={{
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
                mb: 2,
                flexWrap: "wrap",
                gap: 2,
              }}
            >
              <ClickableChips
                selectedTag={selectedTag}
                setSelectedTag={setSelectedTag}
              />

              {isAdmin && (
                <Button
                  component={Link}
                  to="/news/create"
                  variant="contained"
                  color="success"
                  startIcon={<AddIcon />}
                  sx={{
                    borderRadius: 2,
                    px: 2.5,
                    py: 1,
                    fontWeight: 600,
                    textTransform: "none",
                    whiteSpace: "nowrap",
                  }}
                >
                  Beitrag erstellen
                </Button>
              )}
            </Box>

            <Box sx={{ mb: 6 }}>
              <CustomizedInputBase />
            </Box>

            <Grid container spacing={3}>
              {filteredNews.map((item) => (
                <Grid key={item.id} size={{ xs: 12, sm: 6, md: 6 }}>
                  <ReviewCard
                    {...item}
                    isLiked={localLikes.includes(item.id)}
                    onToggleLike={handleToggleLike}
                    onDelete={handleDelete}
                    isAdmin={isAdmin}
                  />
                </Grid>
              ))}
            </Grid>
          </Container>
        </Box>

        {/* Sidebar Rechts (Favoriten) */}
        <Box
          sx={{
            width: { xs: '100%', lg: 350 },
            flexShrink: 0,
            borderLeft: { lg: "1px solid" },
            borderColor: "divider",
            pl: { lg: 3 },
            pt: 4
          }}
        >
          <Typography variant="h5" sx={{ mb: 2, fontWeight: 600 }}>
            Favoriten
          </Typography>
          <Paper variant="outlined" sx={{ borderRadius: 3, overflow: 'hidden' }}>
            <FavoriteList likedIds={localLikes} allNews={allNews} />
          </Paper>
        </Box>
      </Box>
      <Outlet />
    </Sidebar>
  );
}