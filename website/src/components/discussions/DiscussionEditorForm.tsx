import Autocomplete from '@mui/material/Autocomplete';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import MenuItem from '@mui/material/MenuItem';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import type { DtoProgramResponse } from '@lib/api';
import { FORUM_CATEGORIES, FORUM_TAGS } from '@lib/config';

type DiscussionEditorFormProps = {
  programs: DtoProgramResponse[];
  selectedPrograms: DtoProgramResponse[];
  title: string;
  body: string;
  category: string;
  tags: string[];
  eventDate: string;
  location: string;
  isEvent: boolean;
  isStaff: boolean;
  submitting: boolean;
  canSubmit: boolean;
  submitLabel: string;
  onTitleChange: (value: string) => void;
  onBodyChange: (value: string) => void;
  onCategoryChange: (value: string) => void;
  onSelectedProgramsChange: (nextPrograms: DtoProgramResponse[]) => void;
  onTagsChange: (nextTags: string[]) => void;
  onEventDateChange: (value: string) => void;
  onLocationChange: (value: string) => void;
  onSubmit: () => void;
};

const visibleCategories = (isStaff: boolean, selectedCategory: string): string[] => {
  const base = FORUM_CATEGORIES.filter((category) => (category !== 'Ankündigung' && category !== 'Termin') || isStaff);

  if (selectedCategory && !base.some((item) => item === selectedCategory)) {
    return [selectedCategory, ...base];
  }

  return [...base];
};

export function DiscussionEditorForm({
  programs,
  selectedPrograms,
  title,
  body,
  category,
  tags,
  eventDate,
  location,
  isEvent,
  isStaff,
  submitting,
  canSubmit,
  submitLabel,
  onTitleChange,
  onBodyChange,
  onCategoryChange,
  onSelectedProgramsChange,
  onTagsChange,
  onEventDateChange,
  onLocationChange,
  onSubmit,
}: DiscussionEditorFormProps) {
  return (
    <Paper elevation={0} sx={{ p: { xs: 2.5, md: 4 }, borderRadius: 3, border: '1px solid', borderColor: 'divider' }}>
      <Stack spacing={3}>
        <TextField
          label="Titel"
          fullWidth
          value={title}
          onChange={(event) => onTitleChange(event.target.value)}
          required
        />

        <TextField
          select
          label="Kategorie"
          value={category}
          onChange={(event) => onCategoryChange(event.target.value)}
          fullWidth
          required
        >
          {visibleCategories(isStaff, category).map((item) => (
            <MenuItem key={item} value={item}>
              {item}
            </MenuItem>
          ))}
        </TextField>

        <TextField
          label="Inhalt"
          multiline
          rows={6}
          fullWidth
          value={body}
          onChange={(event) => onBodyChange(event.target.value)}
          required
        />

        <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
          <Autocomplete
            multiple
            options={programs}
            getOptionLabel={(option) => option.name || ''}
            value={selectedPrograms}
            isOptionEqualToValue={(option, value) => String(option.id) === String(value.id)}
            onChange={(_, newValue) => onSelectedProgramsChange(newValue)}
            renderInput={(params) => (
              <TextField {...params} label="Studiengänge (Optional)" placeholder="Wählen..." />
            )}
            fullWidth
          />

          <Autocomplete
            multiple
            options={FORUM_TAGS}
            value={tags}
            onChange={(_, newValue) => onTagsChange(newValue)}
            renderTags={(value: readonly string[], getTagProps) =>
              value.map((option: string, index: number) => {
                const { key, ...rest } = getTagProps({ index });
                return <Chip key={key} variant="outlined" label={option} {...rest} />;
              })
            }
            renderInput={(params) => (
              <TextField {...params} label="Tags" placeholder="Wählen..." />
            )}
            fullWidth
          />
        </Stack>

        {isEvent && (
          <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
            <TextField
              label="Event Datum"
              type="datetime-local"
              fullWidth
              value={eventDate}
              onChange={(event) => onEventDateChange(event.target.value)}
              InputLabelProps={{ shrink: true }}
            />
            <TextField
              label="Ort"
              fullWidth
              value={location}
              onChange={(event) => onLocationChange(event.target.value)}
            />
          </Stack>
        )}

        <Box sx={{ display: 'flex', justifyContent: 'flex-end', mt: 2 }}>
          <Button
            variant="contained"
            onClick={onSubmit}
            disabled={submitting || !canSubmit}
            sx={{ borderRadius: 2, px: 4, py: 1.5, fontWeight: 700 }}
          >
            {submitting ? <CircularProgress size={20} color="inherit" /> : submitLabel}
          </Button>
        </Box>
      </Stack>
    </Paper>
  );
}
