import { useTheme } from '@mui/material/styles';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Skeleton from '@mui/material/Skeleton';
import Stack from '@mui/material/Stack';
import { LineChart } from '@mui/x-charts/LineChart';

interface GrowthChartProps {
  title: string;
  value?: string | number;
  trend?: 'up' | 'down' | 'neutral';
  caption?: string;
  data: number[];
  labels: string[];
  loading?: boolean;
}

function AreaGradient({ color, id }: { color: string; id: string }) {
  return (
    <defs>
      <linearGradient id={id} x1="50%" y1="0%" x2="50%" y2="100%">
        <stop offset="0%" stopColor={color} stopOpacity={0.5} />
        <stop offset="100%" stopColor={color} stopOpacity={0} />
      </linearGradient>
    </defs>
  );
}

export default function GrowthChart({ 
  title, 
  value, 
  caption,
  data, 
  labels, 
  loading = false 
}: GrowthChartProps) {
  const theme = useTheme();

  return (
    <Card variant="outlined" sx={{ width: '100%', height: '100%', display: 'flex', flexDirection: 'column', borderRadius: 2, overflow: 'hidden' }}>
      <CardContent sx={{ p: 2, pb: 0, flexGrow: 1, display: 'flex', flexDirection: 'column', '&:last-child': { pb: 0 } }}>
        <Typography component="h2" variant="caption" gutterBottom fontWeight={700} color="text.secondary" sx={{ textTransform: 'uppercase', letterSpacing: '0.05em', fontSize: '0.85rem' }}>
          {title}
        </Typography>
        <Stack sx={{ mb: 1 }}>
          <Stack direction="row" spacing={1} alignItems="center">
            <Typography variant="h4" fontWeight={700} sx={{ fontFamily: 'Space Grotesk, sans-serif' }}>
              {loading ? <Skeleton width={100} /> : (typeof value === 'number' ? value.toLocaleString('de-DE') : value)}
            </Typography>
          </Stack>
          <Typography variant="caption" color="text.secondary" sx={{ fontSize: '0.75rem' }}>{caption}</Typography>
        </Stack>
        <Box sx={{ flexGrow: 1, width: '100%', minHeight: 0, mt: 'auto' }}>
          {loading ? <Skeleton variant="rectangular" height={100} sx={{ borderRadius: 1 }} /> : 
            data.length > 0 ? (
            <LineChart
              series={[{ id: 'growth', label: 'Sitzungen', showMark: false, curve: 'linear', area: true, data: data.length === 1 ? [data[0], data[0]] : data, color: theme.palette.primary.main }]}
              xAxis={[{ scaleType: 'point', data: data.length === 1 ? ['', ''] : labels, tickInterval: (_, i) => (i + 1) % 5 === 0 }]}
              grid={{ horizontal: true }}
              sx={{ '& .MuiAreaElement-series-growth': { fill: "url('#growth-gradient')" }, '& .MuiChartsGrid-line': { strokeDasharray: '4 4', stroke: theme.palette.divider } }}
              hideLegend
              height={320}
            >
              <AreaGradient color={theme.palette.primary.main} id="growth-gradient" />
            </LineChart>
          ) : <Box display="flex" alignItems="center" justifyContent="center" height="100%"><Typography color="textSecondary" variant="body2">Keine Daten verfügbar</Typography></Box>}
        </Box>
      </CardContent>
    </Card>
  );
}
