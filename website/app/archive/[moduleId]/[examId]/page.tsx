import React from 'react';
import ArchiveView from "../../ArchiveView";

export default function ExamArchivePage({ params }: { params: Promise<{ moduleId: string; examId: string }> }) {
  const { moduleId, examId } = React.use(params);
  return <ArchiveView moduleId={moduleId} examId={examId} />;
}
