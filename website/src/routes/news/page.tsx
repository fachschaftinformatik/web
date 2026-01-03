import React, { useState, useEffect } from 'react';
import {
  Card,
  Button,
  CardMedia,
  CardContent,
  CardActions,
  Avatar,
  IconButton,
  Typography,
  List,
  ListItem,
  Divider,
  ListItemText,
  ListItemAvatar,
  Box,
  Chip,
  Stack,
  Paper,
  InputBase,
  CardActionArea
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
import { NEWS_DATA, NewsItem } from '@lib/data';

const TAGS = ["Alle", "Events", "Jobs", "Prüfungen", "Studium", "Sonstiges"];

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

function ClickableChips({ selectedTag, setSelectedTag }: { selectedTag: string, setSelectedTag: (t: string) => void }) {
  return (
    <Stack direction="row" spacing={1} sx={{ overflowX: 'auto', pb: 1 }}>
      {TAGS.map((tag) => (
        <Chip
          key={tag}
          label={tag}
          clickable
          color={selectedTag === tag ? "primary" : "default"}
          onClick={() => setSelectedTag(tag)}
          variant={selectedTag === tag ? "filled" : "outlined"}
        />
      ))}
    </Stack>
  );
}

interface NewsCardProps extends NewsItem {
  isLiked: boolean;
  onToggleLike: (id: number) => void;
  onDelete: (id: number) => void;
  isAdmin: boolean;
}

function NewsCard({ id, title, date, image, summary, isLiked, tags, isNew, isAdmin, onToggleLike, onDelete }: NewsCardProps) {
  const navigate = useNavigate();
  return (
    <Card sx={{ width: '100%', height: '100%', display: 'flex', flexDirection: 'column', borderRadius: 3, boxShadow: 1 }}>
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
                minHeight: '2.4rem',
                fontWeight: 700
              }}
            >
              {title}
            </Typography>
            <Typography variant="caption" sx={{ color: 'text.secondary', display: 'block', mt: 1 }}>
              {date}
            </Typography>
          </Box>
          <Box>
            <Stack direction="row" spacing={1} mb={1} flexWrap="wrap" useFlexGap>
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
          <IconButton onClick={() => onToggleLike(id)} aria-label="add to favorites">
            <FavoriteIcon color={isLiked ? "error" : "inherit"} />
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
            onClick={() => onDelete(id)}
            sx={{ fontSize: "0.75rem" }}
          >
            Löschen
          </Button>
        )}
      </CardActions>
    </Card>
  );
}

function FavoriteList({ likedIds, allNews }: { likedIds: number[], allNews: NewsItem[] }) {
  const favorites = allNews.filter((n) => likedIds.includes(n.id));
  if (favorites.length === 0) return <Typography variant="body2" sx={{ p: 2 }}>Noch keine Favoriten.</Typography>;

  return (
    <List>
      {favorites.map((fav) => (
        <React.Fragment key={fav.id}>
          <ListItem alignItems="flex-start" component={Link} to={`/news/${fav.id}`} sx={{ textDecoration: 'none', color: 'inherit' }}>
            <ListItemAvatar>
              <Avatar variant="rounded" src={fav.image} />
            </ListItemAvatar>
            <ListItemText
              primary={fav.title}
              secondary={fav.date}
              primaryTypographyProps={{ variant: 'body2', fontWeight: 600, noWrap: true }}
              secondaryTypographyProps={{ variant: 'caption' }}
            />
          </ListItem>
          <Divider variant="inset" component="li" />
        </React.Fragment>
      ))}
    </List>
  );
}

export default function NewsRoomPage() {
  const { user } = useAuth();
  const isAdmin = user?.role === "admin" || user?.role === "editor";
  const [selectedTag, setSelectedTag] = useState("Alle");
  const [likedNewsIds, setLikedNewsIds] = useState<number[]>(() => {
    const saved = localStorage.getItem("newsroom-liked");
    return saved ? JSON.parse(saved) : [];
  });
  const [allNews, setAllNews] = useState<NewsItem[]>([]);

  useEffect(() => {
    const custom = JSON.parse(localStorage.getItem("custom-news") || "[]");
    const merged = [...NEWS_DATA, ...custom];

    // Sort logic from main/stashed
    const sorted = merged.sort((a, b) => {
      if (a.isNew && !b.isNew) return -1;
      if (!a.isNew && b.isNew) return 1;
      return b.id - a.id;
    });
    setAllNews(sorted);
  }, []);

  const handleToggleLike = (id: number) => {
    setLikedNewsIds((prev) => {
      const updated = prev.includes(id) ? prev.filter((i) => i !== id) : [...prev, id];
      localStorage.setItem("newsroom-liked", JSON.stringify(updated));
      return updated;
    });
  };

  const handleDelete = (id: number) => {
    setAllNews((prev) => prev.filter(item => item.id !== id));
    // Also remove from custom news in localStorage
    const stored = JSON.parse(localStorage.getItem("custom-news") || "[]");
    const filteredStored = stored.filter((item: any) => item.id !== id);
    localStorage.setItem("custom-news", JSON.stringify(filteredStored));
  };

  const filteredNews = selectedTag === "Alle"
    ? allNews
    : allNews.filter((n) => n.tags?.includes(selectedTag));

  return (
    <Sidebar user={user} title="Newsroom">
      <Box mb={4}>
        <Typography variant="h4" component="h1" fontWeight={700} gutterBottom>
          Newsroom
        </Typography>
        <Typography variant="body1" color="text.secondary">
          Bleibe auf dem Laufenden über alles rund um die Fachschaft und dein Studium.
        </Typography>
      </Box>

      <Box sx={{ display: 'flex', flexDirection: { xs: 'column', lg: 'row' } }}>
        <Box sx={{ flex: 1, mr: { lg: 4 }, mb: 4 }}>
          <Stack
            direction={{ xs: 'column', sm: 'row' }}
            spacing={2}
            sx={{ mb: 4, alignItems: { xs: 'stretch', sm: 'center' } }}
          >
            <Box sx={{ flexShrink: 1 }}>
              <ClickableChips selectedTag={selectedTag} setSelectedTag={setSelectedTag} />
            </Box>
            <Box sx={{ flexGrow: 1, minWidth: { xs: '100%', sm: 280 } }}>
              <CustomizedInputBase />
            </Box>
            <Button
              component={Link}
              to="/news/create"
              variant="contained"
              color="success"
              startIcon={<AddIcon />}
              sx={{
                borderRadius: 2,
                px: 2.5,
                fontWeight: 600,
                textTransform: "none",
                whiteSpace: "nowrap",
                height: 48
              }}
            >
              Erstellen
            </Button>
          </Stack>

          <Grid container spacing={3}>
            {filteredNews.map((item) => (
              <Grid key={item.id} size={{ xs: 12, sm: 6, md: 4 }}>
                <NewsCard
                  {...item}
                  isLiked={likedNewsIds.includes(item.id)}
                  onToggleLike={handleToggleLike}
                  onDelete={handleDelete}
                  isAdmin={isAdmin}
                />
              </Grid>
            ))}
          </Grid>
        </Box>

        <Box
          sx={{
            width: { xs: '100%', lg: 320 },
            flexShrink: 0,
            borderLeft: { lg: "1px solid" },
            borderColor: "divider",
            pl: { lg: 3 },
            pt: { lg: 0, xs: 4 }
          }}
        >
          <Typography variant="h6" sx={{ mb: 2, fontWeight: 700 }}>
            Favoriten
          </Typography>
          <Paper variant="outlined" sx={{ borderRadius: 3, overflow: 'hidden' }}>
            <FavoriteList likedIds={likedNewsIds} allNews={allNews} />
          </Paper>
        </Box>
      </Box>
      <Outlet />
    </Sidebar>
  );
}
