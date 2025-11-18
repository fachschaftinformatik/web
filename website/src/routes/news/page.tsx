import Card from '@mui/material/Card';
import CardMedia from '@mui/material/CardMedia';
import CardContent from '@mui/material/CardContent';
import CardActions from '@mui/material/CardActions';
import Avatar from '@mui/material/Avatar';
import IconButton from '@mui/material/IconButton';
import Typography from '@mui/material/Typography';
import FavoriteIcon from '@mui/icons-material/Favorite';
import StarBorderIcon from '@mui/icons-material/StarBorder';

import Grid from '@mui/material/Grid';
import Container from "@mui/material/Container";

import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import Divider from '@mui/material/Divider';
import ListItemText from '@mui/material/ListItemText';
import ListItemAvatar from '@mui/material/ListItemAvatar';

import Box from "@mui/material/Box";

import Chip from '@mui/material/Chip';
import Stack from '@mui/material/Stack';


import Paper from '@mui/material/Paper';
import InputBase from '@mui/material/InputBase';
import SearchIcon from '@mui/icons-material/Search';
import CardActionArea from '@mui/material/CardActionArea';

import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useEffect } from 'react';

function CustomizedInputBase() {
  return (
    <Paper
      component="form"
      sx={{ p: '2px 4px', display: 'flex', alignItems: 'center', width: "100%" }}
    > 
      <InputBase
        sx={{ ml: 1, flex: 1 }}
        placeholder="Search"
        inputProps={{ 'aria-label': 'search google maps' }}
      />

      <IconButton type="button" sx={{ p: '10px' }} aria-label="search">
        <SearchIcon />
      </IconButton>

      <Divider sx={{ height: 28, m: 0.5 }} orientation="vertical" />
      
    </Paper>
  );
}


// // Tags
//  function ClickableChips() {
//   const handleClick = () => {
//     console.info('You clicked the Chip.');
//   };

//   return (
//     <Stack direction="row" spacing={1}>
//       <Chip label="Alle" onClick={handleClick} />
//       <Chip label="Studium" onClick={handleClick} />
//       <Chip label="Prüfungen" onClick={handleClick} />
//       <Chip label="Events" onClick={handleClick} />
//       <Chip label="Jobs" onClick={handleClick} />
//     </Stack>
//   );
// }


function ClickableChips({ selectedTag, setSelectedTag }: { selectedTag: string, setSelectedTag: (tag: string) => void }) {
  const tags = ["Alle", "Studium", "Prüfungen", "Events", "Jobs"];

  return (
    <Stack direction="row" spacing={1}>
      {tags.map((tag) => (
        <Chip
          key={tag}
          label={tag}
          onClick={() => setSelectedTag(tag)}
          color={selectedTag === tag ? "success" : "default"}
          size='medium'
        />
      ))}
    </Stack>
  );
}


// Typ für News-Einträge
type NewsItem = {
  id: number;
  title: string;
  date: string;
  image: string;
  summary: string;
  content: string;
  links:string[];
  isNew: boolean;
  tags: string[];};

//  Array mit allen News
export const newsDaten: NewsItem[] = [
  {
    id: 1,
    title: "6-Year-Old Horse Dies at Belmont Park After Race Injury",
    date: "September 14, 2016",
    image: "https://images.unsplash.com/photo-1607746882042-944635dfe10e?auto=format&fit=crop&w=400&h=200&q=80",
    summary: "A tragic accident occurred during the race, raising concerns about animal safety...",
    content:"Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.",
    links:[
      "https://conference2025.com",
      "https://more-info.com"
    ],
    isNew : true,
    tags: ["Events", "Jobs"]

  },
  {
    id: 2,
    title: "Tech Conference 2025 Kicks Off in Berlin",
    date: "September 12, 2025",
    image: "https://images.unsplash.com/photo-1607746882042-944635dfe10e?auto=format&fit=crop&w=400&h=200&q=80",
    summary: "Industry leaders gather to discuss the future of AI, robotics, and quantum computing...",
    content:"",
    links:[
      "https://conference2025.com",
      "https://more-info.com"
    ],
    isNew : true,
    tags: ["Prüfungen", "Studium"]

  },
  {
    id: 3,
    title: "New Climate Agreement Signed",
    date: "September 10, 2025",
    image: "https://images.unsplash.com/photo-1607746882042-944635dfe10e?auto=format&fit=crop&w=400&h=200&q=80",
    summary: "World leaders agreed on a historic pact aiming to reduce emissions worldwide...",
    content:"",
    links:[
      "https://conference2025.com",
      "https://more-info.com"
    ],
    isNew: true,
    tags: ["Events", "Studium"]

  },
  {
    id: 4,
    title: "-Year-Old Horse Dies at Belmont Park After Race Injury",
    date: "September 8, 2025",
    image: "https://images.unsplash.com/photo-1607746882042-944635dfe10e?auto=format&fit=crop&w=400&h=200&q=80",
    summary: "Scientists reported a promising new therapy with encouraging early trial results...",
    content:"",
    links:[
      "https://conference2025.com",
      "https://more-info.com"
    ],
    isNew: false,
    tags: ["Studium"]

  },
  {
    id: 5,
    title: "Breakthrough in Cancer Research",
    date: "September 8, 2025",
    image: "https://images.unsplash.com/photo-1607746882042-944635dfe10e?auto=format&fit=crop&w=400&h=200&q=80",
    summary: "Scientists reported a promising new therapy with encouraging early trial results...",
    content:"",
    links:[
      "https://conference2025.com",
      "https://more-info.com"
    ],
    isNew: true,
    tags: ["Events", "Studium"]

  },
  {
    id: 6,
    title: "Breakthrough in Cancer Research",
    date: "September 8, 2025",
    image: "https://images.unsplash.com/photo-1607746882042-944635dfe10e?auto=format&fit=crop&w=400&h=200&q=80",
    summary: "Scientists reported a promising new therapy with encouraging early trial results...",
    content:"Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.",
    links:[
      "https://conference2025.com",
      "https://more-info.com"
    ],
    isNew: false,
    tags: ["Events", "Studium"]

  },
  {
    id: 7,
    title: "Breakthrough in Cancer Research",
    date: "September 8, 2025",
    image: "https://images.unsplash.com/photo-1607746882042-944635dfe10e?auto=format&fit=crop&w=400&h=200&q=80",
    summary: "Scientists reported a promising new therapy with encouraging early trial results...",
    content:"Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.",
    links:[
      "https://conference2025.com",
      "https://more-info.com"
    ],
    isNew: true,
    tags: ["Events", "Studium"]

  }
];


 // Props für Card
type Props = {
  id: number;
  title: string;
  date: string;
  image: string;
  summary: string;
  isNew: boolean;
  isLiked: boolean;
  onToggleLike: (id: number) => void;
};


// die Karten mit den News Inhalten
function RecipeReviewCard({ id, title, date, image, summary, isNew, isLiked, onToggleLike }: Props) {
  //um zur Detailseite zu Navigieren
   const navigate = useNavigate();
  return (
    <Card sx={{ width: '100%',height: '100%', display: 'flex', flexDirection: 'column' , borderRadius: 3}}>
      <CardActionArea sx={{ flexGrow: 1 }}  onClick={() => navigate(`/news/${id}`)}>
        <CardMedia
          component="img"
          height="200"
          image={image}
          alt={title}
        />

        <CardContent sx={{ display: 'flex',flexDirection: 'column', height: '100%'  }}>
          <Box sx={{  height: '5.2rem', mb: 1 }}>
            <Typography variant="h6"
                sx={{
                display: '-webkit-box',
                WebkitLineClamp: 2,
                WebkitBoxOrient: 'vertical',
                overflow: 'hidden',
                lineHeight: '1.2rem',
              }}
            >
              {title}
            </Typography>
            <Typography variant="caption" sx={{ color: 'text.secondary', display: 'block', mt:1, minHeight: '3rem',}}>
              {date}
            </Typography>
              
          </Box>
           
          <Box>
            {/* Themen-Tags */}
            <Stack direction="row" spacing={1} mb={1}>
              <Chip label="Events" size="small" />
              <Chip label="Studium" size="small" />
            </Stack>
            {/* Beschreibung */}
            <Typography variant="body2" sx={{ color: 'text.secondary' }}>
              {summary}
            </Typography>
          </Box>
        </CardContent>
      </CardActionArea>
      {/* 🔹 Like (& Share) Buttons */}
      <CardActions disableSpacing>
        <IconButton onClick={() => onToggleLike(id)} aria-label="add to favorites">
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
      </CardActions>
    </Card>
  );
}
// damit die funktion LikedIn.in liste hinzufügen funktioniert
type FavoriteListProps = {
  likedIds: number[];
};
function FavoriteList({ likedIds }: FavoriteListProps) {
  const likedItems = newsDaten.filter((item) => likedIds.includes(item.id));
  return (
    <List sx={{ width: '100%', maxWidth: 350 }}>
      {likedItems.map((item) => (
        <ListItem key={item.id}>
          <ListItemAvatar>
            <Avatar src={item.image} variant="square" sx={{ width: 85, height: 70, pr: 2 }}/>
          </ListItemAvatar>
          <ListItemText
      primary={
    <Link
      to={`/news/${item.id}`}
      style={{ textDecoration: 'none', color: 'inherit' }}
    >
      <Typography variant="body2" sx={{
        fontWeight: 'bold',
        '&:hover': { textDecoration: 'underline', cursor: 'pointer' }
      }}>
        {item.title}
      </Typography>
      <Typography variant='caption' sx={{ color: 'text.secondary' }} >
        {item.date}
      </Typography>
    </Link>
  }
/>
        </ListItem>
      ))}
    </List>
  );
}






  export default function NewsLayout() {
    const [selectedTag, setSelectedTag] = useState<string>("Alle");

    // 1. State für Likes
   const [likedNewsIds, setLikedNewsIds] = useState<number[]>([]);

   // 2. Effekt: beim ersten Laden schauen, ob Likes im localStorage sind
   useEffect(() => {
     const savedLikes = localStorage.getItem('likedNews');
     if (savedLikes) {
       setLikedNewsIds(JSON.parse(savedLikes));
     }
   }, []);

  const handleToggleLike = (id: number) => {
   setLikedNewsIds((prev) => {
     const updatedLikes = prev.includes(id)
       ? prev.filter((itemId) => itemId !== id)
       : [...prev, id];

     localStorage.setItem('likedNews', JSON.stringify(updatedLikes));
     return updatedLikes;
   });
 };


  // Sortierte News: neue zuerst
  const sortedNews = [...newsDaten].sort((a, b) => {
    if (a.isNew && !b.isNew) return -1;
    if (!a.isNew && b.isNew) return 1;
    return 0;
  });

  const filteredNews = sortedNews.filter((item) =>
  selectedTag === "Alle" ? true : item.tags.includes(selectedTag)
);


  return (
    <Box sx={{ display: 'flex', flexDirection: 'row', px: 2, pl: 30}}>
      {/* 🔹 Hauptinhalt */}
      <Box sx={{ flex: 1, maxWidth: "1000px", pr: 8 }}>
        <Container sx={{ pt: '200px', pb: 2 }}>

          {/* Suche + Chips */}
          <Stack
            direction={{ xs: 'column', sm: 'row' }}
            spacing={2}
            sx={{ mb: 2, alignItems: { xs: 'stretch', sm: 'center' } }}
          >
            <Box sx={{ flexShrink: 1 }}>
              {/* <ClickableChips /> */}
              <ClickableChips selectedTag={selectedTag} setSelectedTag={setSelectedTag} />
            </Box>
            <Box sx={{ flexGrow: 1, minWidth: { xs: '100%', sm: 280 } }}>
              <CustomizedInputBase />
            </Box>
          </Stack>
          {/* Cards im Grid */}
          <Grid container spacing={5}>
            {filteredNews.map((item) => (
             <Grid key={item.id} size={{ xs: 12, sm: 6 }}>
                <RecipeReviewCard
                  title={item.title}
                  date={item.date}
                  image={item.image}
                  summary={item.summary}
                  isNew={item.isNew}
                  id={item.id}
                  isLiked={likedNewsIds.includes(item.id)}
                  onToggleLike={handleToggleLike}
                />
              </Grid>
            ))}
          </Grid>
        </Container>
      </Box>


      

      {/* Favoriten Liste */}
      <Box 
        sx={{
          width: 370,
          height: '100vh',
          overflowY: 'auto',
          border: "7px solid #e0e0e0",
          borderRadius: 5,
          p: 1,
          mt: "200px"
        }}
      >

        <Typography variant="h3" sx={{ mb: 1 }}>
          Favoriten
        </Typography>

        <FavoriteList likedIds={likedNewsIds} />
      </Box>

    </Box>
  );
}
