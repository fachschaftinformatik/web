import React from 'react';
import {
  Box,
  Breadcrumbs,
  Button,
  Chip,
  Container,
  Divider,
  Link,
  List,
  ListItem,
  Paper,
  Stack,
  Typography,
  Pagination,
} from '@mui/material';
import { Link as RouterLink } from 'react-router-dom';
import { mockForumPosts } from './forumData';
import SidebarLayout from '../layouts/SidebarLayout';
import { useAuth } from '../AuthContext';

const POSTS_PER_PAGE = 5;

const ForumPostsPage: React.FC = () => {
  const { user } = useAuth();
  const [page, setPage] = React.useState(1);
  const pageCount = Math.max(1, Math.ceil(mockForumPosts.length / POSTS_PER_PAGE));

  const paginatedPosts = React.useMemo(() => {
    const startIndex = (page - 1) * POSTS_PER_PAGE;
    return mockForumPosts.slice(startIndex, startIndex + POSTS_PER_PAGE);
  }, [page]);

  React.useEffect(() => {
    if (page > pageCount) {
      setPage(pageCount);
    }
  }, [page, pageCount]);

  if (!user) {
    return null;
  }

  return (
    <SidebarLayout user={user} title="Forum">
      <Container maxWidth="md" sx={{ py: 4 }}>
        <Breadcrumbs aria-label="breadcrumb trail" sx={{ mb: 2 }}>
          <Link component={RouterLink} underline="hover" color="inherit" to="/dashboard">
            Dashboard
          </Link>
          <Typography color="text.primary">Forum</Typography>
        </Breadcrumbs>

        <Stack direction="row" spacing={2} alignItems="center" justifyContent="space-between" mb={2}>
          <Box>
            <Typography variant="h4" fontWeight={600}>
              Alle Forenbeiträge
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Blättere durch alle Beiträge dieses Profils. Pagination sorgt für Übersicht bei vielen Posts.
            </Typography>
          </Box>
          <Button variant="contained" component={RouterLink} to="/dashboard">
            Zurück
          </Button>
        </Stack>

        <Paper variant="outlined">
          <List disablePadding>
            {paginatedPosts.map((post, index) => (
              <React.Fragment key={post.id}>
                {index > 0 && <Divider component="li" />}
                <ListItem
                  alignItems="flex-start"
                  sx={{
                    px: { xs: 2, sm: 3 },
                    py: 2,
                    flexDirection: 'column',
                    gap: 1,
                  }}
                >
                  <Stack direction="row" spacing={1} alignItems="center" width="100%">
                    <Typography variant="h6" flex={1}>
                      {post.title}
                    </Typography>
                    <Chip label={post.category} size="small" variant="outlined" />
                  </Stack>
                  <Typography variant="body2" color="text.secondary" width="100%">
                    {post.excerpt}
                  </Typography>
                  <Typography variant="caption" color="text.secondary" width="100%">
                    {new Date(post.createdAt).toLocaleString()} · {post.replies} Antworten
                  </Typography>
                  <Box width="100%" display="flex" justifyContent="flex-end">
                    <Button variant="text" size="small" disabled>
                      Zum Beitrag (in Arbeit)
                    </Button>
                  </Box>
                </ListItem>
              </React.Fragment>
            ))}
          </List>
          <Box display="flex" justifyContent="center" py={2}>
            <Pagination
              count={pageCount}
              page={page}
              onChange={(_, value) => setPage(value)}
              color="primary"
              showFirstButton
              showLastButton
            />
          </Box>
        </Paper>
      </Container>
    </SidebarLayout>
  );
};

export default ForumPostsPage;
