import { useTheme } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Skeleton from '@mui/material/Skeleton';
import { SparkLineChart } from '@mui/x-charts/SparkLineChart';
import { areaElementClasses } from '@mui/x-charts/LineChart';


export type StatCardProps = {
  title: string;
  value: string | number;
  interval?: string;
  trend?: 'up' | 'down' | 'neutral';
  data?: number[];
  loading?: boolean;
};

function AreaGradient({ color, id }: { color: string; id: string }) {
  return (
    <defs>
      <linearGradient id={id} x1="50%" y1="0%" x2="50%" y2="100%">
        <stop offset="0%" stopColor={color} stopOpacity={0.3} />
        <stop offset="100%" stopColor={color} stopOpacity={0} />
      </linearGradient>
    </defs>
  );
}

export default function StatCard({
  title,
  value,
  interval = 'Letzte 90 Tage',
  trend = 'neutral',
  data = [],
  loading = false,
}: StatCardProps) {
  const theme = useTheme();

  const trendColors = {
    up: theme.palette.success.main,
    down: theme.palette.error.main,
    neutral: theme.palette.grey[500],
  };

  const chartColor = trendColors[trend];

  const id = `area-gradient-${title.replace(/\s+/g, '-').toLowerCase()}`;

  return (
    <Card variant="outlined" sx={{ height: '100%', width: '100%', borderRadius: 2, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <CardContent sx={{ p: 2, flexGrow: 1, display: 'flex', flexDirection: 'column', '&:last-child': { pb: 2 } }}>
        <Typography component="h2" variant="caption" gutterBottom fontWeight={700} sx={{ color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.05em', mb: 0.5, fontSize: '0.85rem' }}>
          {title}
        </Typography>
        <Stack
          direction="column"
          sx={{ justifyContent: 'space-between', flexGrow: 1, minHeight: 0, gap: 1 }}
        >
          <Box>
            <Stack
              direction="row"
              sx={{ justifyContent: 'space-between', alignItems: 'center' }}
            >
              <Typography variant="h5" component="p" fontWeight={700} sx={{ fontFamily: 'Space Grotesk, sans-serif' }}>
                {loading ? <Skeleton width={60} /> : typeof value === 'number' ? value.toLocaleString('de-DE') : value}
              </Typography>
            </Stack>

            <Typography variant="caption" sx={{ color: 'text.secondary', fontSize: '0.8rem', display: 'block' }}>
              {interval}
            </Typography>
          </Box>
          <Box sx={{ width: '100%', mt: 'auto', height: 80 }}>
            {data.length > 0 ? (
              <SparkLineChart
                color={chartColor}
                data={data.length === 1 ? [data[0], data[0]] : data}
                area
                height={80}
                showHighlight
                showTooltip
                sx={{
                  [`& .${areaElementClasses.root}`]: {
                    fill: `url(#${id})`,
                  },
                }}
              >
                <AreaGradient color={chartColor} id={id} />
              </SparkLineChart>

            ) : (
              <Box sx={{ height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <Skeleton variant="rectangular" width="100%" height={30} sx={{ borderRadius: 1, opacity: 0.1 }} />
              </Box>
            )}
          </Box>
        </Stack>
      </CardContent>
    </Card>
  );

}
