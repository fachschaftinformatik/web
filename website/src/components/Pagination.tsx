import Box from '@mui/material/Box';
import MuiPagination from '@mui/material/Pagination';
import { SxProps, Theme } from '@mui/material/styles';

interface PaginationProps {
  count: number;
  page: number;
  onChange: (event: React.ChangeEvent<unknown>, value: number) => void;
  disabled?: boolean;
  sx?: SxProps<Theme>;
}

export default function Pagination({ count, page, onChange, disabled, sx }: PaginationProps) {
  if (count <= 1) return null;
  return (
    <Box display="flex" justifyContent="center" mt={4} sx={sx}>
      <MuiPagination
        count={count}
        page={page}
        onChange={onChange}
        color="primary"
        shape="rounded"
        variant="outlined"
        disabled={disabled}
      />
    </Box>
  );
}
