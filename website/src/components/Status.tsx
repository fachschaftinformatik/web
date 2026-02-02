import { useEffect, useState } from 'react';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import CircularProgress from '@mui/material/CircularProgress';
import { useTheme, alpha } from '@mui/material/styles';
import MeetingRoomRounded from '@mui/icons-material/MeetingRoomRounded';
import { getOfficeStatus, putOfficeStatus } from '@lib/api';
import type { DtoUserResponse as User } from '@lib/api/types.gen';

interface StatusProps {
  user?: User | null;
}

export default function Status({ user }: StatusProps) {
  const [occupied, setOccupied] = useState<boolean>(false);
  const [loading, setLoading] = useState<boolean>(true);
  const [updating, setUpdating] = useState<boolean>(false);
  const theme = useTheme();

  const isAuthorized = user?.role === 'admin' || user?.role === 'editor';

  const fetchStatus = async () => {
    try {
      const { data } = await getOfficeStatus();
      if (data) {
        setOccupied(data.occupied || false);
      }
    } catch (err) {
      console.error('Failed to fetch office status', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStatus();
    const interval = setInterval(fetchStatus, 120000);
    return () => clearInterval(interval);
  }, []);

  const handleToggle = async () => {
    if (!isAuthorized || updating) return;
    const newValue = !occupied;
    setUpdating(true);
    try {
      await putOfficeStatus({
        body: { occupied: newValue }
      });
      setOccupied(newValue);
    } catch (err) {
      console.error('Failed to update office status', err);
    } finally {
      setUpdating(false);
    }
  };

  if (loading && !isAuthorized) return null;

  const bg = occupied
    ? alpha(theme.palette.success.main, 0.1)
    : alpha(theme.palette.text.secondary, 0.05);
  
  const border = occupied
    ? alpha(theme.palette.success.main, 0.2)
    : alpha(theme.palette.divider, 0.1);
  
  const color = occupied
    ? theme.palette.success.main
    : theme.palette.text.secondary;

  return (
    <Box
      sx={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        py: 0.75,
        px: 2,
        bgcolor: bg,
        borderBottom: '1px solid',
        borderColor: border,
        transition: 'all 0.3s ease',
      }}
    >
      <Box
        onClick={isAuthorized ? handleToggle : undefined}
        sx={{
          display: 'flex',
          alignItems: 'center',
          gap: 1.5,
          width: '100%',
          maxWidth: 'lg',
          px: { xs: 2.5, md: 5 },
          cursor: isAuthorized ? 'pointer' : 'default',
          '&:hover': isAuthorized ? {
            opacity: 0.8,
          } : {},
          transition: 'opacity 0.2s ease',
        }}
      >
        <MeetingRoomRounded sx={{ color, fontSize: 22 }} />
        <Typography
          variant="body1"
          sx={{
            fontWeight: 700,
            color: color,
            flexGrow: 1,
            fontFamily: '"Space Grotesk", sans-serif',
            letterSpacing: '0.02em',
            lineHeight: 1.2,
            textTransform: 'uppercase',
            fontSize: '0.95rem',
            userSelect: 'none',
          }}
        >
          {occupied ? 'Büro besetzt' : 'Büro nicht besetzt'}
        </Typography>
        {updating && (
          <CircularProgress size={16} color="inherit" sx={{ opacity: 0.5, ml: 1 }} />
        )}
      </Box>
    </Box>
  );
}
