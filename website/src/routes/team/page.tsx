import React from "react";
import {
  Container,
  Typography,
  Card,
  CardMedia,
  CardContent,
  Box,
  Link,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Stack,
  TextField,
  IconButton,
  Tooltip,
  InputAdornment,
} from "@mui/material";
import AccountCircleIcon from "@mui/icons-material/AccountCircle";
import EditRoundedIcon from "@mui/icons-material/EditRounded";
import AddRoundedIcon from "@mui/icons-material/AddRounded";
import DeleteOutlineRoundedIcon from "@mui/icons-material/DeleteOutlineRounded";
import SaveRoundedIcon from "@mui/icons-material/SaveRounded";
import CloseRoundedIcon from "@mui/icons-material/CloseRounded";
import { useLocation } from "react-router-dom";
import { Sidebar } from "@components/layout";
import { useAuth } from "@lib/auth";
import { teamSections, TeamMember, TeamSection } from "@lib/data";

const ADMIN_PREVIEW_KEY = "team_admin_preview";
const EMAIL_DOMAIN = "@fsv-whs.de";

const deepCloneSections = (sections: TeamSection[]) =>
  sections.map((s) => ({
    ...s,
    members: s.members.map((m) => ({ ...m })),
  }));

const getLocalPart = (email: string) => {
  if (!email) return "";
  const atIndex = email.indexOf("@");
  return atIndex > -1 ? email.slice(0, atIndex) : email;
};

const normalizeLocalPart = (value: string) => value.split("@")[0].trim();

export default function Team() {
  const { user } = useAuth();
  const location = useLocation();
  const [adminPreview, setAdminPreview] = React.useState(false);
  const canEdit = adminPreview || user?.role === "admin" || user?.role === "editor";
  const [sections, setSections] = React.useState<TeamSection[]>(teamSections);
  const [editOpen, setEditOpen] = React.useState(false);
  const [draft, setDraft] = React.useState<TeamSection[]>([]);
  const [uploadOpen, setUploadOpen] = React.useState(false);
  const [uploadTarget, setUploadTarget] = React.useState<{
    sectionId: string;
    memberId: number;
  } | null>(null);
  const [deleteTarget, setDeleteTarget] = React.useState<{
    sectionId: string;
    memberId: number;
  } | null>(null);

  const updateMemberImage = React.useCallback(
    (sectionId: string, memberId: number, img: string | null) => {
      setDraft((prev) =>
        prev.map((s) =>
          s.id === sectionId
            ? {
              ...s,
              members: s.members.map((m) =>
                m.id === memberId ? { ...m, img } : m
              ),
            }
            : s
        )
      );
    },
    []
  );

  const readFileAsDataUrl = React.useCallback((file: File) => {
    return new Promise<string>((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = () => resolve(reader.result as string);
      reader.onerror = () => reject(new Error("Datei konnte nicht gelesen werden."));
      reader.readAsDataURL(file);
    });
  }, []);

  const handleImageFile = React.useCallback(
    async (file: File, sectionId: string, memberId: number) => {
      try {
        const dataUrl = await readFileAsDataUrl(file);
        updateMemberImage(sectionId, memberId, dataUrl);
      } catch {
        void 0;
      }
    },
    [readFileAsDataUrl, updateMemberImage]
  );

  const openUploadDialog = (sectionId: string, memberId: number) => {
    setUploadTarget({ sectionId, memberId });
    setUploadOpen(true);
  };

  const closeUploadDialog = () => {
    setUploadOpen(false);
    setUploadTarget(null);
  };

  const confirmDeleteMember = () => {
    if (!deleteTarget) return;
    setDraft((prev) =>
      prev.map((s) =>
        s.id === deleteTarget.sectionId
          ? {
            ...s,
            members: s.members.filter((m) => m.id !== deleteTarget.memberId),
          }
          : s
      )
    );
    setDeleteTarget(null);
  };

  React.useEffect(() => {
    const searchParams = new URLSearchParams(location.search);
    const param = searchParams.get("admin");
    if (param === "1") {
      setAdminPreview(true);
      if (typeof window !== "undefined") {
        window.localStorage.setItem(ADMIN_PREVIEW_KEY, "1");
      }
      return;
    }
    if (param === "0") {
      setAdminPreview(false);
      if (typeof window !== "undefined") {
        window.localStorage.removeItem(ADMIN_PREVIEW_KEY);
      }
      return;
    }
    if (typeof window !== "undefined") {
      setAdminPreview(window.localStorage.getItem(ADMIN_PREVIEW_KEY) === "1");
    }
  }, [location.search]);

  const Section = ({ title, members }: { title: string; members: TeamMember[] }) => (
    <Box sx={{ my: 4 }}>
      <Typography variant="h5" gutterBottom>{title}</Typography>
      <Box
        sx={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fit, minmax(260px, 320px))",
          justifyContent: "flex-start",
          gap: 4,
        }}
      >
        {members.map((m) => (
          <Card key={m.id} sx={{ borderRadius: 2, overflow: "hidden", height: "100%" }}>

            {m.img ? (
              <CardMedia
                component="img"
                height="320"
                image={m.img}
                alt={m.name}
                sx={{ objectFit: "contain", backgroundColor: "grey.100" }}
              />
            ) : (
              <Box
                sx={{
                  height: 320,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  backgroundColor: 'grey.200',
                }}
              >
                <AccountCircleIcon
                  sx={{
                    fontSize: 200,
                    color: 'grey.400'
                  }}
                />
              </Box>
            )}
            <CardContent>
              <Typography variant="subtitle1">{m.name}</Typography>
              {m.email && (
                <Link
                  href={`mailto:${m.email}`}
                  variant="body2"
                  color="text.secondary"
                  underline="hover"
                >
                  {m.email}
                </Link>
              )}
            </CardContent>
          </Card>
        ))}
      </Box>
    </Box>
  );
  return (
    <Sidebar user={user} title="Team">
      <Container sx={{ py: 4 }}>
        {canEdit && (
          <Box sx={{ display: "flex", justifyContent: "flex-end", mb: 2 }}>
            <Button
              variant="contained"
              startIcon={<EditRoundedIcon />}
              onClick={() => {
                setDraft(deepCloneSections(sections));
                setEditOpen(true);
              }}
              sx={{ textTransform: "none", borderRadius: 2 }}
              disableElevation
            >
              Bearbeiten
            </Button>
          </Box>
        )}
        <Typography variant="h4" align="center" gutterBottom>
          FSV-Team
        </Typography>
        {sections.map((section) => (
          <Section key={section.id} title={section.title} members={section.members} />
        ))}
      </Container>

      <Dialog open={editOpen} onClose={() => setEditOpen(false)} fullWidth maxWidth="md">
        <DialogTitle sx={{ display: "flex", alignItems: "center", gap: 1 }}>
          <EditRoundedIcon fontSize="small" />
          Team bearbeiten
        </DialogTitle>
        <DialogContent dividers>
          <Stack spacing={3}>
            {draft.map((section) => (
              <Box key={section.id} sx={{ border: "1px solid", borderColor: "divider", borderRadius: 2, p: 2 }}>
                <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ mb: 1 }}>
                  <Typography variant="h6">{section.title}</Typography>
                  <Button
                    size="small"
                    startIcon={<AddRoundedIcon />}
                    onClick={() => {
                      setDraft((prev) => {
                        const maxId = Math.max(0, ...prev.flatMap((s) => s.members.map((m) => m.id)));
                        return prev.map((s) =>
                          s.id === section.id
                            ? {
                              ...s,
                              members: [
                                ...s.members,
                                { id: maxId + 1, name: "", email: "", img: null },
                              ],
                            }
                            : s
                        );
                      });
                    }}
                  >
                    Mitglied hinzufügen
                  </Button>
                </Stack>

                <Stack spacing={2}>
                  {section.members.map((member) => (
                    <Box
                      key={member.id}
                      sx={{
                        p: 2,
                        borderRadius: 2,
                        border: "1px solid",
                        borderColor: "divider",
                        bgcolor: "background.paper",
                      }}
                    >
                      <Stack direction={{ xs: "column", sm: "row" }} spacing={2} alignItems="flex-start">
                        <Box sx={{ width: { xs: "100%", sm: 240 } }}>
                          <TextField
                            id={`name-${member.id}`}
                            name="name"
                            label="Name"
                            value={member.name}
                            onChange={(e) =>
                              setDraft((prev) =>
                                prev.map((s) =>
                                  s.id === section.id
                                    ? {
                                      ...s,
                                      members: s.members.map((m) =>
                                        m.id === member.id ? { ...m, name: e.target.value } : m
                                      ),
                                    }
                                    : s
                                )
                              )
                            }
                            fullWidth
                          />
                        </Box>
                        <Box sx={{ flex: 1, minWidth: { xs: "100%", sm: 280 } }}>
                          <TextField
                            id={`email-${member.id}`}
                            name="email"
                            label="E-Mail"
                            value={getLocalPart(member.email)}
                            onChange={(e) =>
                              setDraft((prev) =>
                                prev.map((s) =>
                                  s.id === section.id
                                    ? {
                                      ...s,
                                      members: s.members.map((m) =>
                                        m.id === member.id
                                          ? {
                                            ...m,
                                            email: normalizeLocalPart(e.target.value)
                                              ? `${normalizeLocalPart(e.target.value)}${EMAIL_DOMAIN}`
                                              : "",
                                          }
                                          : m
                                      ),
                                    }
                                    : s
                                )
                              )
                            }
                            fullWidth
                            InputProps={{
                              endAdornment: (
                                <InputAdornment position="end">{EMAIL_DOMAIN}</InputAdornment>
                              ),
                            }}
                          />
                        </Box>
                        <Stack spacing={1} sx={{ minWidth: { xs: "100%", sm: 140 } }}>
                          <Button
                            variant="contained"
                            onClick={() => openUploadDialog(section.id, member.id)}
                            color="success"
                            sx={{ textTransform: "none" }}
                            disableElevation
                          >
                            Bild hochladen
                          </Button>
                          <Typography variant="body2" color="text.secondary">
                            {member.img ? "Bild gesetzt" : "Kein Bild"}
                          </Typography>
                        </Stack>
                        <Tooltip title="Entfernen">
                          <IconButton
                            color="error"
                            onClick={() => setDeleteTarget({ sectionId: section.id, memberId: member.id })}
                            sx={{ alignSelf: "center" }}
                          >
                            <DeleteOutlineRoundedIcon />
                          </IconButton>
                        </Tooltip>
                      </Stack>
                    </Box>
                  ))}
                  {section.members.length === 0 && (
                    <Typography variant="body2" color="text.secondary">
                      Keine Mitglieder eingetragen.
                    </Typography>
                  )}
                </Stack>
              </Box>
            ))}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button
            startIcon={<CloseRoundedIcon />}
            onClick={() => setEditOpen(false)}
            color="inherit"
            sx={{ textTransform: "none" }}
          >
            Abbrechen
          </Button>
          <Button
            startIcon={<SaveRoundedIcon />}
            variant="contained"
            onClick={() => {
              setSections(draft);
              setEditOpen(false);
            }}
            sx={{ textTransform: "none" }}
          >
            Speichern
          </Button>
        </DialogActions>
      </Dialog>
      <Dialog open={uploadOpen} onClose={closeUploadDialog} fullWidth maxWidth="sm">
        <DialogTitle>Bild hochladen</DialogTitle>
        <DialogContent dividers>
          <input
            accept="image/*"
            id="team-img-upload"
            type="file"
            hidden
            onChange={async (e) => {
              const file = e.target.files?.[0];
              if (!file || !uploadTarget) return;
              await handleImageFile(file, uploadTarget.sectionId, uploadTarget.memberId);
              e.target.value = "";
              closeUploadDialog();
            }}
          />
          <Box
            onDragOver={(e) => {
              e.preventDefault();
            }}
            onDrop={(e) => {
              e.preventDefault();
              const file = e.dataTransfer.files?.[0];
              if (file && uploadTarget) {
                handleImageFile(file, uploadTarget.sectionId, uploadTarget.memberId);
                closeUploadDialog();
              }
            }}
            sx={{
              border: "2px dashed",
              borderColor: "divider",
              borderRadius: 2,
              p: 4,
              textAlign: "center",
              bgcolor: "background.default",
            }}
          >
            <Typography variant="body1" sx={{ mb: 1 }}>
              Bild hier ablegen
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
              oder Datei auswählen
            </Typography>
            <label htmlFor="team-img-upload">
              <Button variant="contained" component="span" sx={{ textTransform: "none" }}>
                Datei durchsuchen
              </Button>
            </label>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={closeUploadDialog} color="inherit" sx={{ textTransform: "none" }}>
            Abbrechen
          </Button>
        </DialogActions>
      </Dialog>
      <Dialog open={Boolean(deleteTarget)} onClose={() => setDeleteTarget(null)} maxWidth="xs" fullWidth>
        <DialogTitle>Daten löschen</DialogTitle>
        <DialogContent dividers>
          <Typography variant="body2" color="text.secondary">
            Die Daten werden unwiderruflich gelöscht. Trotzdem fortfahren?
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteTarget(null)} color="inherit" sx={{ textTransform: "none" }}>
            Abbrechen
          </Button>
          <Button onClick={confirmDeleteMember} color="error" variant="contained" sx={{ textTransform: "none" }}>
            Löschen
          </Button>
        </DialogActions>
      </Dialog>
    </Sidebar>
  );
}
