import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import { SxProps, Theme } from '@mui/material/styles';
import ArrowBackRounded from '@mui/icons-material/ArrowBackRounded';
import { useNavigate } from 'react-router-dom';

interface BackProps {
  to: string | -1;
  label?: string;
  sx?: SxProps<Theme>;
  icon?: boolean;
}

export default function Back({ to, label = "Zurück", sx, icon = false }: BackProps) {
  const navigate = useNavigate();
  const onClick = () => (to === -1 ? navigate(-1) : navigate(to));

  if (icon) {
    return (
      <IconButton onClick={onClick} sx={{ bgcolor: 'action.hover', ...sx }}>
        <ArrowBackRounded />
      </IconButton>
    );
  }

  return (
    <Button
      startIcon={<ArrowBackRounded />}
      onClick={onClick}
      sx={{
        mb: 2,
        color: 'text.secondary',
        textTransform: 'none',
        '&:hover': { color: 'primary.main', bgcolor: 'transparent', textDecoration: 'underline' },
        ...sx
      }}
    >
      {label}
    </Button>
  );
}
