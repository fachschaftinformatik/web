import * as React from 'react';
import { PieChart } from '@mui/x-charts/PieChart';
import { useDrawingArea } from '@mui/x-charts/hooks';
import { styled } from '@mui/material/styles';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Box from '@mui/material/Box';
import Skeleton from '@mui/material/Skeleton';
import Stack from '@mui/material/Stack';

interface StyledTextProps {
  variant: 'primary' | 'secondary';
}

const StyledText = styled('text', {
  shouldForwardProp: (prop) => prop !== 'variant',
})<StyledTextProps>(({ theme }) => ({
  textAnchor: 'middle',
  dominantBaseline: 'central',
  fill: theme.palette.text.secondary,
  variants: [
    {
      props: { variant: 'primary' },
      style: { fontSize: theme.typography.h5.fontSize, fontWeight: theme.typography.h5.fontWeight },
    },
    {
      props: ({ variant }) => variant !== 'primary',
      style: { fontSize: theme.typography.body2.fontSize, fontWeight: theme.typography.body2.fontWeight },
    },
  ],
}));

interface PieCenterLabelProps {
  primaryText: string;
  secondaryText: string;
}

function PieCenterLabel({ primaryText, secondaryText }: PieCenterLabelProps) {
  const { width, height, left, top } = useDrawingArea();
  return (
    <React.Fragment>
      <StyledText variant="primary" x={left + width / 2} y={top + height / 2 - 10}>{primaryText}</StyledText>
      <StyledText variant="secondary" x={left + width / 2} y={top + height / 2 + 14}>{secondaryText}</StyledText>
    </React.Fragment>
  );
}

interface ProgramDistributionChartProps {
  title: string;
  data: { label: string; value: number }[];
  loading?: boolean;
}

export default function ProgramDistributionChart({ title, data, loading = false }: ProgramDistributionChartProps) {
  const total = data.reduce((acc, curr) => acc + curr.value, 0);
  const colors = ['#0288d1', '#7b1fa2', '#2e7d32', '#ed6c02', '#d32f2f', '#00838f', '#455a64', '#fbc02d'];
  const pieData = data.map((item, index) => ({ id: index, label: item.label, value: item.value }));

  return (
    <Card variant="outlined" sx={{ height: '100%', width: '100%', borderRadius: 2, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <CardContent sx={{ p: 2, pb: 0, flexGrow: 1, display: 'flex', flexDirection: 'column', '&:last-child': { pb: 0 } }}>
        <Typography component="h2" variant="caption" gutterBottom fontWeight={700} color="text.secondary" sx={{ textTransform: 'uppercase', letterSpacing: '0.05em', fontSize: '0.85rem' }}>
          {title}
        </Typography>
        <Box sx={{ flexGrow: 1, width: '100%', minHeight: 0, display: 'flex', flexDirection: { xs: 'column', md: 'row' }, alignItems: 'center', justifyContent: 'center', mt: 'auto' }}>
          {loading ? <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', flexGrow: 1 }}><Skeleton variant="circular" width={160} height={160} /></Box> : 
            data.length > 0 ? (
            <>
              <Box sx={{ flexGrow: 1, height: '100%', minWidth: 0, display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
                <PieChart colors={colors} series={[{ data: pieData, innerRadius: '70%', outerRadius: '95%', paddingAngle: 3, cornerRadius: 6, highlightScope: { fade: 'global', highlight: 'item' }, cx: '50%', cy: '50%' }]} height={300} hideLegend>
                  <PieCenterLabel primaryText={total.toLocaleString('de-DE')} secondaryText="Gesamt" />
                </PieChart>
              </Box>
              <Box sx={{ width: { xs: '100%', md: '350px' }, maxHeight: '100%', overflowY: 'auto', pr: 2, py: 1 }}>
                <Stack spacing={1.5}>
                  {data.map((item, index) => (
                    <Box key={index} sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                      <Stack direction="row" spacing={1.5} alignItems="center" sx={{ minWidth: 0 }}>
                        <Box sx={{ width: 12, height: 12, borderRadius: '50%', bgcolor: colors[index % colors.length], flexShrink: 0 }} />
                        <Typography variant="body2" sx={{ fontSize: '0.9rem', fontWeight: 600, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', color: 'text.primary' }}>{item.label}</Typography>
                      </Stack>
                      <Typography variant="body2" color="text.secondary" sx={{ fontSize: '0.9rem', fontWeight: 700, ml: 2, flexShrink: 0 }}>{item.value.toLocaleString('de-DE')}</Typography>
                    </Box>
                  ))}
                </Stack>
              </Box>
            </>
          ) : <Box sx={{ flexGrow: 1, display: 'flex', alignItems: 'center', justifyContent: 'center' }}><Typography variant="body2" color="textSecondary">Keine Daten verfügbar</Typography></Box>}
        </Box>
      </CardContent>
    </Card>
  );
}
