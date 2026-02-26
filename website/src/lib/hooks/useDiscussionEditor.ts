import { useEffect, useState } from 'react';
import { getPrograms, postDiscussions, type DtoProgramResponse } from '@lib/api';
import { getDiscussionById, updateDiscussionPost } from '@lib/adapters/discussions';
import { FORUM_CATEGORIES } from '@lib/config';
import { toPostId } from '@lib/types/domain';
import { toErrorMessage } from '@lib/types/guards';

type DiscussionType = 'discussion' | 'news' | 'event';

const getTypeFromCategory = (category: string): DiscussionType => {
  if (category === 'Ankündigung') {
    return 'news';
  }

  if (category === 'Termin') {
    return 'event';
  }

  return 'discussion';
};

const getFallbackCategoryFromType = (type: string | undefined): string => {
  if (type === 'news') return 'Ankündigung';
  if (type === 'event') return 'Termin';
  return '';
};

const getCategoryFromPost = (tags: string[], type: string | undefined): string =>
  tags.find((tag) => FORUM_CATEGORIES.some((category) => category === tag)) || getFallbackCategoryFromType(type);

type UseDiscussionEditorOptions = {
  mode: 'create' | 'edit';
  postIdParam?: string;
};

type DiscussionDraft = {
  selectedPrograms: DtoProgramResponse[];
  title: string;
  body: string;
  category: string;
  tags: string[];
  eventDate: string;
  location: string;
};

const EMPTY_DRAFT: DiscussionDraft = {
  selectedPrograms: [],
  title: '',
  body: '',
  category: '',
  tags: [],
  eventDate: '',
  location: '',
};

export const useDiscussionEditor = ({ mode, postIdParam }: UseDiscussionEditorOptions) => {
  const [programs, setPrograms] = useState<DtoProgramResponse[]>([]);
  const [draft, setDraft] = useState<DiscussionDraft>(EMPTY_DRAFT);
  const [loading, setLoading] = useState(mode === 'edit');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  const isEditMode = mode === 'edit';
  const postId = isEditMode ? toPostId(postIdParam) : null;

  const type = getTypeFromCategory(draft.category);
  const isEvent = type === 'event';
  const canSubmit = draft.title.trim().length > 0 && draft.body.trim().length > 0 && draft.category.length > 0;

  useEffect(() => {
    const loadPrograms = async () => {
      const { data } = await getPrograms();
      if (data) {
        setPrograms(data);
      }
    };

    void loadPrograms();
  }, []);

  useEffect(() => {
    if (!isEditMode) {
      return;
    }

    if (!postId) {
      setLoading(false);
      setError('Ungueltige Beitrags-ID.');
      return;
    }

    const loadPost = async () => {
      setLoading(true);
      setError('');

      try {
        const { data, error: apiError } = await getDiscussionById(postId);
        if (apiError || !data) {
          setError('Beitrag konnte nicht geladen werden.');
          return;
        }

        const loadedTags = data.tags || [];
        const loadedCategory = getCategoryFromPost(loadedTags, data.type);

        setDraft({
          title: data.title || '',
          body: data.body || '',
          category: loadedCategory,
          tags: loadedTags.filter((tag) => tag !== loadedCategory),
          eventDate: data.event_date || '',
          location: data.location || '',
          selectedPrograms: (data.programs || []).map((program) => ({
            id: program.id,
            name: program.name,
          })),
        });
      } catch {
        setError('Fehler beim Laden.');
      } finally {
        setLoading(false);
      }
    };

    void loadPost();
  }, [isEditMode, postId]);

  const submit = async () => {
    if (!canSubmit) {
      return false;
    }

    setSubmitting(true);
    setError('');

    const payload = {
      title: draft.title,
      body: draft.body,
      type,
      programs: draft.selectedPrograms.map((program) => String(program.id)),
      tags: [draft.category, ...draft.tags],
      event_date: draft.eventDate || undefined,
      location: draft.location || undefined,
    };

    try {
      if (!isEditMode) {
        const { error: apiError } = await postDiscussions({ body: payload });
        if (apiError) {
          throw new Error(toErrorMessage(apiError) || 'Fehler beim Erstellen.');
        }
      } else {
        if (!postId) {
          throw new Error('Ungueltige Beitrags-ID.');
        }

        const { error: apiError } = await updateDiscussionPost(postId, payload);
        if (apiError) {
          throw new Error(toErrorMessage(apiError) || 'Fehler beim Speichern.');
        }
      }

      return true;
    } catch (submitError: unknown) {
      setError(toErrorMessage(submitError) || 'Ein unerwarteter Fehler ist aufgetreten.');
      return false;
    } finally {
      setSubmitting(false);
    }
  };

  const updateDraft = <K extends keyof DiscussionDraft>(key: K, value: DiscussionDraft[K]) => {
    setDraft((previous) => ({ ...previous, [key]: value }));
  };

  return {
    programs,
    selectedPrograms: draft.selectedPrograms,
    title: draft.title,
    body: draft.body,
    category: draft.category,
    tags: draft.tags,
    eventDate: draft.eventDate,
    location: draft.location,
    isEvent,
    loading,
    submitting,
    error,
    canSubmit,
    setTitle: (value: string) => updateDraft('title', value),
    setBody: (value: string) => updateDraft('body', value),
    setCategory: (value: string) => updateDraft('category', value),
    setTags: (value: string[]) => updateDraft('tags', value),
    setEventDate: (value: string) => updateDraft('eventDate', value),
    setLocation: (value: string) => updateDraft('location', value),
    setSelectedPrograms: (value: DtoProgramResponse[]) => updateDraft('selectedPrograms', value),
    submit,
  };
};
