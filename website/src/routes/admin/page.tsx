import React, { useEffect, useState } from 'react';
import {
  Typography,
  Box,
  useTheme,
} from '@mui/material';

import { getAdminDashboard, type DtoAdminDashboardResponse } from '@lib/api';
import { Sidebar } from '@components/layout';
import { useAuth } from '@lib/auth';

import StatCard from '@components/StatCard';
import UserGrowthChart from '@components/UserGrowthChart';
import ActivityTrendChart from '@components/ActivityTrendChart';
import ProgramDistributionChart from '@components/ProgramDistributionChart';

const AdminDashboard: React.FC = () => {
  const { user } = useAuth();
  const [data, setData] = useState<DtoAdminDashboardResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const theme = useTheme();

  const fetchDashboardData = async () => {
    setLoading(true);
    try {
      const { data, error } = await getAdminDashboard();
      if (error) throw error;
      setData(data || null);
    } catch (err) {
      console.error('Failed to fetch admin dashboard data:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDashboardData();
  }, []);

  const growthData = data?.user_growth_trend?.map(item => Number(item.count) || 0) || [];
  const growthLabels = data?.user_growth_trend?.map(item => item.date ? new Date(item.date).toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit' }) : '') || [];

  const sessionData = data?.session_trend?.map(item => Number(item.count) || 0) || [];
  const sessionLabels = data?.session_trend?.map(item => item.date ? new Date(item.date).toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit' }) : '') || [];

  const examData = data?.exam_growth_trend?.map(item => Number(item.count) || 0) || [];
  const discData = data?.discussion_growth_trend?.map(item => Number(item.count) || 0) || [];
  const moduleGrowthData = data?.module_growth_trend?.map(item => Number(item.count) || 0) || [];
  const programGrowthData = data?.program_growth_trend?.map(item => Number(item.count) || 0) || [];

  const activityData = data?.activity_trend?.map(item => Number(item.count) || 0) || [];
  const activityLabels = data?.activity_trend?.map(item => item.date ? new Date(item.date).toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit' }) : '') || [];

  const distributionData = data?.program_distribution?.map((item) => ({
    label: item.name || 'Unbekannt',
    value: Number(item.value) || 0,
  })) || [];

  const GRID_ASPECT_RATIO = '1.4 / 1'; 

  return (
    <Sidebar user={user} title="Übersicht" maxWidth="lg">
      <Box sx={{ display: 'flex', flexDirection: 'column', width: '100%' }}>
        <Box sx={{ mb: 3 }}>
          <Typography variant="h4" component="h1" fontWeight={700} sx={{ fontFamily: 'Space Grotesk, sans-serif' }}>
            Übersicht
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Metriken und Statistiken der letzten 90 Tage.
          </Typography>
        </Box>

        <Box sx={{ 
          display: 'grid', 
          gridTemplateColumns: { xs: '1fr', sm: 'repeat(2, 1fr)', md: 'repeat(4, 1fr)' },
          gap: 2,
          width: '100%'
        }}>
          <Box sx={{ aspectRatio: { sm: GRID_ASPECT_RATIO }, display: 'flex', minWidth: 0 }}>
            <StatCard
              title="Benutzer"
              value={data?.stats?.user_count || 0}
              data={growthData}
              trend="up"
              loading={loading}
            />
          </Box>
          <Box sx={{ aspectRatio: { sm: GRID_ASPECT_RATIO }, display: 'flex', minWidth: 0 }}>
            <StatCard
              title="Prüfungen"
              value={data?.stats?.archive_count || 0}
              data={examData}
              trend="neutral"
              loading={loading}
            />
          </Box>
          <Box sx={{ aspectRatio: { sm: GRID_ASPECT_RATIO }, display: 'flex', minWidth: 0 }}>
            <StatCard
              title="Beiträge"
              value={data?.stats?.post_count || 0}
              data={discData}
              trend="up"
              loading={loading}
            />
          </Box>
          <Box sx={{ aspectRatio: { sm: GRID_ASPECT_RATIO }, display: 'flex', minWidth: 0 }}>
            <StatCard
              title="Aktivitäten"
              value={data?.stats?.activity_count || 0}
              data={activityData}
              trend="down"
              loading={loading}
            />
          </Box>

          <Box sx={{ 
            gridColumn: { xs: '1', sm: 'span 2', md: 'span 2' }, 
            aspectRatio: { sm: GRID_ASPECT_RATIO }, 
            display: 'flex', 
            minWidth: 0 
          }}>
            <UserGrowthChart 
              title="Sitzungen" 
              value={data?.stats?.session_count || 0}
              trend="up"
              caption="Sitzungen pro Tag in den letzten 90 Tagen"
              data={sessionData} 
              labels={sessionLabels} 
              loading={loading} 
            />
          </Box>
          <Box sx={{ 
            gridColumn: { xs: '1', sm: 'span 2', md: 'span 2' }, 
            aspectRatio: { sm: GRID_ASPECT_RATIO }, 
            display: 'flex', 
            minWidth: 0 
          }}>
            <ActivityTrendChart 
              title="Aktivitäten" 
              value={data?.stats?.activity_count || 0}
              trend="down"
              caption="Aktivitäten pro Tag in den letzten 90 Tagen"
              data={activityData} 
              labels={activityLabels} 
              loading={loading} 
            />
          </Box>

          <Box sx={{ 
            gridColumn: { xs: '1', sm: 'span 3', md: 'span 3' }, 
            gridRow: { md: 'span 2' },
            aspectRatio: { sm: '1.95 / 1' }, 
            display: 'flex', 
            minWidth: 0 
          }}>
            <ProgramDistributionChart 
              title="Studiengänge" 
              data={distributionData} 
              loading={loading} 
            />
          </Box>

          <Box sx={{ aspectRatio: { sm: GRID_ASPECT_RATIO }, display: 'flex', minWidth: 0 }}>
            <StatCard
              title="Module"
              value={data?.stats?.module_count || 0}
              data={moduleGrowthData}
              loading={loading}
            />
          </Box>
          <Box sx={{ aspectRatio: { sm: GRID_ASPECT_RATIO }, display: 'flex', minWidth: 0 }}>
            <StatCard
              title="Studiengänge"
              value={data?.stats?.program_count || 0}
              data={programGrowthData}
              loading={loading}
            />
          </Box>
        </Box>
      </Box>
    </Sidebar>
  );
};

export default AdminDashboard;
