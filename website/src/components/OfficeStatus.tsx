import React, { useEffect, useState } from 'react';
import {
  Box,
  Typography,
  Switch,
  FormControlLabel,
  alpha,
  useTheme,
  CircularProgress,
} from '@mui/material';
import MeetingRoomRounded from '@mui/icons-material/MeetingRoomRounded';
import { getOfficeStatus, putOfficeStatus } from '@lib/api';
import type { DtoUserResponse as User } from '@lib/api/types.gen';

interface OfficeStatusProps {
  user?: User | null;
}

const OfficeStatus = ({ user }: OfficeStatusProps) => {
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
    // Refresh status every 2 minutes
    const interval = setInterval(fetchStatus, 120000);
    return () => clearInterval(interval);
  }, []);

  const handleToggle = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = event.target.checked;
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
        sx={{
          display: 'flex',
          alignItems: 'center',
          gap: 1.5,
          width: '100%',
          maxWidth: 'lg',
          px: { xs: 2.5, md: 5 },
        }}
      >
        <MeetingRoomRounded sx={{ color, fontSize: 20 }} />
        <Typography
          variant="body2"
          sx={{
            fontWeight: 800,
            color: color,
            flexGrow: 1,
            fontFamily: '"Space Grotesk", sans-serif',
            letterSpacing: '0.02em',
            lineHeight: 1.2,
            textTransform: 'uppercase',
          }}
        >
          {occupied ? 'Büro besetzt' : 'Büro nicht besetzt'}
        </Typography>
          {isAuthorized && (
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
              {updating && <CircularProgress size={16} color="inherit" sx={{ opacity: 0.5 }} />}
              <FormControlLabel
                control={
                  <Switch
                    size="small"
                    checked={occupied}
                    onChange={handleToggle}
                    disabled={updating}
                    color="success"
                  />
                }
                label={
                  <Typography variant="caption" sx={{ fontWeight: 600, color: 'text.secondary', display: { xs: 'none', sm: 'block' } }}>
                    Status ändern
                  </Typography>
                }
                labelPlacement="start"
                sx={{ m: 0 }}
              />
            </Box>
          )}
        </Box>
      </Box>
  );
};

export default OfficeStatus;
