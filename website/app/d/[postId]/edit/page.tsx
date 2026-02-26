'use client';

import React from "react";
import { useRouter } from "next/navigation";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Alert from "@mui/material/Alert";
import CircularProgress from "@mui/material/CircularProgress";
import IconButton from "@mui/material/IconButton";
import ArrowBackRounded from "@mui/icons-material/ArrowBackRounded";

import { useSessionUser } from "@lib/hooks/useSessionUser";
import { Sidebar } from "@components/layout";
import { useDiscussionEditor } from "@lib/hooks/useDiscussionEditor";
import { DiscussionEditorForm } from "@components/discussions/DiscussionEditorForm";

export default function EditPost({ params }: { params: Promise<{ postId: string }> }) {
    const { postId } = React.use(params);
    const { user, isStaff } = useSessionUser();
    const router = useRouter();

    const {
        programs,
        selectedPrograms,
        title,
        body,
        category,
        tags,
        eventDate,
        location,
        isEvent,
        loading,
        submitting,
        error,
        canSubmit,
        setTitle,
        setBody,
        setCategory,
        setTags,
        setEventDate,
        setLocation,
        setSelectedPrograms,
        submit,
    } = useDiscussionEditor({ mode: "edit", postIdParam: postId });


    const handleSubmit = async () => {
        const success = await submit();
        if (success) {
            router.push("/discussions");
        }
    };


    if (loading) {
        return (
            <Sidebar user={user} title="Beitrag bearbeiten">
                <Box display="flex" justifyContent="center" p={4}><CircularProgress /></Box>
            </Sidebar>
        );
    }

    return (
        <Sidebar user={user} title="Beitrag bearbeiten" maxWidth="md">
            <Box>
                <Box sx={{ mb: 4, display: 'flex', alignItems: 'center', gap: 2 }}>
                    <IconButton onClick={() => router.push("/discussions")} sx={{ bgcolor: 'action.hover' }}>
                        <ArrowBackRounded />
                    </IconButton>
                    <Typography variant="h4" fontWeight={700} sx={{ fontSize: { xs: '1.5rem', md: '2.125rem' } }}>
                        Beitrag bearbeiten
                    </Typography>
                </Box>

                {error && <Alert severity="error" sx={{ mb: 3 }}>{error}</Alert>}

                <DiscussionEditorForm
                    programs={programs}
                    selectedPrograms={selectedPrograms}
                    title={title}
                    body={body}
                    category={category}
                    tags={tags}
                    eventDate={eventDate}
                    location={location}
                    isEvent={isEvent}
                    isStaff={isStaff}
                    submitting={submitting}
                    canSubmit={canSubmit}
                    submitLabel="Speichern"
                    onTitleChange={setTitle}
                    onBodyChange={setBody}
                    onCategoryChange={setCategory}
                    onSelectedProgramsChange={setSelectedPrograms}
                    onTagsChange={setTags}
                    onEventDateChange={setEventDate}
                    onLocationChange={setLocation}
                    onSubmit={handleSubmit}
                />
            </Box>
        </Sidebar>
    );
}
