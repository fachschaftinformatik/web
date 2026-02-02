import React, { createContext, useContext, useEffect } from 'react';
import Container from '@mui/material/Container';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';

interface LayoutContextValue {
  setTitle: (title: string) => void;
  setHeaderActions: (actions: React.ReactNode) => void;
}

const LayoutContext = createContext<LayoutContextValue | undefined>(undefined);

export const useLayout = () => {
  const context = useContext(LayoutContext);
  if (!context) throw new Error('useLayout must be used within Layout');
  return context;
};

export const LayoutContextProvider: React.FC<{ children: React.ReactNode, value: LayoutContextValue }> = ({ children, value }) => (
  <LayoutContext.Provider value={value}>{children}</LayoutContext.Provider>
);

interface PageProps {
  children: React.ReactNode;
  title?: string;
  description?: string;
  hideHeader?: boolean;
  headerActions?: React.ReactNode;
  fullBleed?: boolean;
  maxWidth?: "xs" | "sm" | "md" | "lg" | "xl" | false;
}

export default function Page({
  children,
  title,
  description,
  hideHeader = false,
  headerActions,
  fullBleed = false,
  maxWidth = "lg"
}: PageProps) {
  const { setTitle, setHeaderActions } = useLayout();

  useEffect(() => {
    if (title) {
      setTitle(title);
      document.title = `${title} | FSV Informatik`;
    }
    setHeaderActions(headerActions || null);
  }, [title, headerActions, setTitle, setHeaderActions]);

  return (
    <Container 
      maxWidth={fullBleed ? false : maxWidth} 
      disableGutters 
      sx={{ flexGrow: 1, p: fullBleed ? 0 : { xs: 2.5, md: 3 } }}
    >
      {!hideHeader && title && (
        <Box sx={{ mb: 4 }}>
          <Typography variant="h4" fontWeight={700} gutterBottom>{title}</Typography>
          {description && <Typography variant="body1" color="text.secondary">{description}</Typography>}
        </Box>
      )}
      {children}
    </Container>
  );
}
