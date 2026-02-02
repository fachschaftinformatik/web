import { useState, useEffect } from 'react';
import Box from '@mui/material/Box';
import IconButton from '@mui/material/IconButton';
import Autocomplete from '@mui/material/Autocomplete';
import CircularProgress from '@mui/material/CircularProgress';
import InputBase from '@mui/material/InputBase';
import Avatar from '@mui/material/Avatar';
import ListItem from '@mui/material/ListItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import useMediaQuery from '@mui/material/useMediaQuery';
import { useTheme, alpha, styled } from '@mui/material/styles';
import SearchIcon from '@mui/icons-material/Search';
import DashboardRounded from '@mui/icons-material/DashboardRounded';
import QuestionAnswerRounded from '@mui/icons-material/QuestionAnswerRounded';
import SchoolRounded from '@mui/icons-material/SchoolRounded';
import CloseRounded from '@mui/icons-material/CloseRounded';
import { useNavigate } from 'react-router-dom';
import { client } from '@lib/api/client.gen';
import { getAvatarUrl } from '@lib/images';

const SearchContainer = styled('div')(({ theme }) => ({
  position: 'relative',
  borderRadius: theme.shape.borderRadius,
  backgroundColor: alpha(theme.palette.common.white, 0.15),
  '&:hover': { backgroundColor: alpha(theme.palette.common.white, 0.25) },
  marginLeft: 0,
  width: '100%',
  [theme.breakpoints.up('sm')]: { marginLeft: theme.spacing(1), width: 'auto' },
}));

const SearchIconWrapper = styled('div')(({ theme }) => ({
  padding: theme.spacing(0, 2), height: '100%', position: 'absolute', pointerEvents: 'none',
  display: 'flex', alignItems: 'center', justifyContent: 'center',
}));

const StyledInputBase = styled(InputBase)(({ theme }) => ({
  color: 'inherit', width: '100%',
  '& .MuiInputBase-input': {
    padding: theme.spacing(1, 1, 1, 0), paddingLeft: `calc(1em + ${theme.spacing(4)})`,
    transition: theme.transitions.create('width'),
    [theme.breakpoints.up('sm')]: { width: '12ch', '&:focus': { width: '20ch' } },
  },
}));

interface SearchResult {
  type: 'archive' | 'module' | 'user' | 'discussion';
  id: string;
  title: string;
  subtitle: string;
  url: string;
}

export default function Search() {
  const theme = useTheme();
  const isSmUp = useMediaQuery(theme.breakpoints.up('sm'));
  const [mobileOpen, setMobileOpen] = useState(false);
  const [open, setOpen] = useState(false);
  const [options, setOptions] = useState<readonly SearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [inputValue, setInputValue] = useState('');
  const navigate = useNavigate();
  const isSearchOpen = open && inputValue.length >= 2;

  useEffect(() => {
    if (isSearchOpen) document.body.style.overflow = 'hidden';
    else document.body.style.overflow = '';
    return () => { document.body.style.overflow = ''; };
  }, [isSearchOpen]);

  useEffect(() => {
    let active = true;
    if (inputValue.length < 2) {
      void Promise.resolve().then(() => { setOptions([]); setLoading(false); });
      return undefined;
    }
    void Promise.resolve().then(() => setLoading(true));
    const timer = setTimeout(() => {
      client.request({ method: 'GET', url: '/search', query: { q: inputValue } }).then(({ data }) => {
        if (active && data) setOptions(data as SearchResult[]);
      }).finally(() => { if (active) setLoading(false); });
    }, 400);
    return () => { active = false; clearTimeout(timer); };
  }, [inputValue]);

  if (!isSmUp && !mobileOpen) return <IconButton color="inherit" onClick={() => setMobileOpen(true)}><SearchIcon /></IconButton>;

  return (
    <Box sx={!isSmUp && mobileOpen ? { position: 'absolute', top: 0, left: 0, right: 0, bottom: 0, bgcolor: alpha(theme.palette.background.paper, 0.9), backdropFilter: 'blur(10px)', display: 'flex', alignItems: 'center', px: 1.5, zIndex: 1200 } : { mr: 1 }}>
      <Autocomplete
        id="global-search" fullWidth open={open && inputValue.length >= 2} onOpen={() => setOpen(true)} onClose={() => setOpen(false)}
        isOptionEqualToValue={(option, value) => String(option.id) === String(value.id)} getOptionLabel={(option) => option.title}
        filterOptions={(x) => x} options={options} loading={loading} value={null}
        onInputChange={(_, newInputValue) => setInputValue(newInputValue)}
        onChange={(_, value) => { if (value) { navigate(value.url); setInputValue(''); setOptions([]); if (!isSmUp) setMobileOpen(false); } }}
        noOptionsText="Keine Ergebnisse" loadingText="Wird geladen..." clearIcon={null} popupIcon={null}
        sx={{ '& .MuiAutocomplete-inputRoot': { flexWrap: 'nowrap' } }}
        renderInput={(params) => (
          <SearchContainer ref={params.InputProps.ref} sx={!isSmUp ? { mx: 0, flex: 1, backgroundColor: alpha(theme.palette.text.primary, 0.05), color: theme.palette.text.primary, '&:hover': { backgroundColor: alpha(theme.palette.text.primary, 0.08) }, '& .MuiInputBase-input': { color: theme.palette.text.primary, '&::placeholder': { color: theme.palette.text.secondary, opacity: 1 } }, '& .MuiSvgIcon-root': { color: theme.palette.primary.main } } : {}}>
            <SearchIconWrapper>{loading ? <CircularProgress color="inherit" size={20} /> : <SearchIcon />}</SearchIconWrapper>
            <StyledInputBase inputProps={params.inputProps} placeholder="Suchen…" autoFocus={!isSmUp} />
          </SearchContainer>
        )}
        renderOption={(props, option) => (
          <ListItem {...props} key={String(option.id) + option.type}>
            <ListItemIcon sx={{ minWidth: 40 }}>
              {option.type === 'module' ? <DashboardRounded fontSize="small" /> :
                option.type === 'user' ? <Avatar src={getAvatarUrl(`/api/v1/auth/avatars/${option.id}/generated_v4.svg`)} sx={{ width: 24, height: 24 }} /> :
                  option.type === 'discussion' ? <QuestionAnswerRounded fontSize="small" /> : <SchoolRounded fontSize="small" />}
            </ListItemIcon>
            <ListItemText primary={option.title} secondary={option.subtitle} primaryTypographyProps={{ variant: 'body2', fontWeight: 600 }} secondaryTypographyProps={{ variant: 'caption' }} />
          </ListItem>
        )}
      />
      {!isSmUp && <IconButton color="inherit" onClick={() => setMobileOpen(false)} sx={{ ml: 1 }}><CloseRounded /></IconButton>}
    </Box>
  );
}
