import { useEffect, useState } from 'react';
import Box from '@mui/material/Box';
import { getAdminDashboard, type DtoAdminDashboardResponse } from '@lib/api';
import Page from '@components/Page';
import StatCard from '@components/StatCard';
import UserGrowthChart from '@components/UserGrowthChart';
import ActivityTrendChart from '@components/ActivityTrendChart';
import ProgramDistributionChart from '@components/ProgramDistributionChart';

export default function Admin() {
  const [data, setData] = useState<DtoAdminDashboardResponse | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getAdminDashboard().then(({ data }) => {
      setData(data || null);
      setLoading(false);
    });
  }, []);

  const s = data?.stats;
  const trend = (arr?: { count?: number | string }[]) => arr?.map(i => Number(i.count) || 0) || [];
  const labels = (arr?: { date?: string }[]) => arr?.map(i => i.date ? new Date(i.date).toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit' }) : '') || [];

  return (
    <Page title="Übersicht" description="Metriken der letzten 90 Tage.">
      <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: 'repeat(2, 1fr)', md: 'repeat(4, 1fr)' }, gap: 2 }}>
        <StatCard title="Benutzer" value={s?.user_count || 0} data={trend(data?.user_growth_trend)} trend="up" loading={loading} />
        <StatCard title="Prüfungen" value={s?.archive_count || 0} data={trend(data?.exam_growth_trend)} trend="neutral" loading={loading} />
        <StatCard title="Beiträge" value={s?.post_count || 0} data={trend(data?.discussion_growth_trend)} trend="up" loading={loading} />
        <StatCard title="Aktivitäten" value={s?.activity_count || 0} data={trend(data?.activity_trend)} trend="down" loading={loading} />

        <Box sx={{ gridColumn: { sm: 'span 2' }, aspectRatio: '1.4/1' }}>
          <UserGrowthChart title="Sitzungen" value={s?.session_count || 0} data={trend(data?.session_trend)} labels={labels(data?.session_trend)} loading={loading} />
        </Box>
        <Box sx={{ gridColumn: { sm: 'span 2' }, aspectRatio: '1.4/1' }}>
          <ActivityTrendChart title="Aktivitäten" value={s?.activity_count || 0} data={trend(data?.activity_trend)} labels={labels(data?.activity_trend)} loading={loading} />
        </Box>

        <Box sx={{ gridColumn: { sm: 'span 3' } }}>
          <ProgramDistributionChart title="Studiengänge" data={data?.program_distribution?.map(i => ({ label: i.name || '?', value: Number(i.value) || 0 })) || []} loading={loading} />
        </Box>
        <Box sx={{ display: 'grid', gap: 2 }}>
          <StatCard title="Module" value={s?.module_count || 0} data={trend(data?.module_growth_trend)} loading={loading} />
          <StatCard title="Studiengänge" value={s?.program_count || 0} data={trend(data?.program_growth_trend)} loading={loading} />
        </Box>
      </Box>
    </Page>
  );
}
