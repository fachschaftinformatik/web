import React from 'react';
import ArchiveView from '../ArchiveView';

export default function ModuleArchivePage({ params }: { params: Promise<{ moduleId: string }> }) {
  const { moduleId } = React.use(params);
  return <ArchiveView moduleId={moduleId} />;
}
