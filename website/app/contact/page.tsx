'use client';

import Typography from '@mui/material/Typography';
import Paper from '@mui/material/Paper';
import Box from '@mui/material/Box';
import Divider from '@mui/material/Divider';
import { Sidebar } from '@components/layout';
import { useAuth } from '@lib/auth';


export default function ContactPage() {
    const { user } = useAuth();

    return (
        <Sidebar user={user} title="Kontakt" maxWidth="lg">
            <Box>
                <Box sx={{ mb: 4 }}>
                    <Typography variant="h4" fontWeight={700} gutterBottom>
                        Kontakt
                    </Typography>
                    <Typography variant="body1" color="text.secondary">
                        Informationen zum Verein, rechtliche Hinweise und wie du uns erreichen kannst.
                    </Typography>
                </Box>

                <Paper elevation={0} sx={{ p: 4, borderRadius: 3, border: '1px solid', borderColor: 'divider' }}>
                    <Typography variant="h6" gutterBottom>
                        Impressum
                    </Typography>
                    <Divider sx={{ my: 2 }} />
                    <Typography variant="body2" paragraph>
                        Angaben gemäß § 5 TMG:
                    </Typography>
                    <Typography variant="body2" fontWeight={600}>
                        Fachschaftsvertretung Informatik
                    </Typography>
                    <Typography variant="body2" paragraph>
                        Neidenburgerstr. 43<br />
                        45897 Gelsenkirchen
                    </Typography>

                    <Typography variant="body2" fontWeight={600} gutterBottom>
                        Vertreten durch:
                    </Typography>
                    <Typography variant="body2" paragraph>
                        Lennart Möller
                    </Typography>

                    <Typography variant="body2" fontWeight={600} gutterBottom>
                        Kontakt:
                    </Typography>
                    <Typography variant="body2" paragraph>
                        Telefon: (02 09) 95 96-416<br />
                        E-Mail: <a href="mailto:kontakt@fachschaftinformatik.de" style={{ color: 'inherit' }}>kontakt@fachschaftinformatik.de</a>
                    </Typography>

                    <Typography variant="body2" fontWeight={600} gutterBottom>
                        Verantwortlich für den Inhalt nach § 55 Abs. 2 RStV:
                    </Typography>
                    <Typography variant="body2" paragraph>
                        Eric König, Stellv. Vorsitz<br />
                        Julia Rabenhorst, Öffentlichkeit<br />
                        Toni Gudaqi, Öffentlichkeit
                    </Typography>

                    <Typography variant="body2" fontWeight={600} gutterBottom sx={{ mt: 3 }}>
                        Streitschlichtung
                    </Typography>
                    <Typography variant="body2" paragraph>
                        Wir sind nicht bereit oder verpflichtet, an Streitbeilegungsverfahren vor einer Verbraucherschlichtungsstelle teilzunehmen.
                    </Typography>

                    <Typography variant="body2" fontWeight={600} gutterBottom sx={{ mt: 3 }}>
                        Haftung für Inhalte
                    </Typography>
                    <Typography variant="body2" paragraph>
                        Als Diensteanbieter sind wir gemäß § 7 Abs.1 TMG für eigene Inhalte auf diesen Seiten nach den allgemeinen Gesetzen verantwortlich. Nach §§ 8 bis 10 TMG sind wir als Diensteanbieter jedoch nicht verpflichtet, übermittelte oder gespeicherte fremde Informationen zu überwachen oder nach Umständen zu forschen, die auf eine rechtswidrige Tätigkeit hinweisen.
                    </Typography>

                    <Typography variant="body2" fontWeight={600} gutterBottom sx={{ mt: 3 }}>
                        Haftung für Links
                    </Typography>
                    <Typography variant="body2" paragraph>
                        Unser Angebot enthält Links zu externen Websites Dritter, auf deren Inhalte wir keinen Einfluss haben. Deshalb können wir für diese fremden Inhalte auch keine Gewähr übernehmen. Für die Inhalte der verlinkten Seiten ist stets der jeweilige Anbieter oder Betreiber der Seiten verantwortlich.
                    </Typography>

                    <Typography variant="body2" fontWeight={600} gutterBottom sx={{ mt: 3 }}>
                        Urheberrecht
                    </Typography>
                    <Typography variant="body2" paragraph>
                        Die durch die Seitenbetreiber erstellten Inhalte und Werke auf diesen Seiten unterliegen dem deutschen Urheberrecht. Die Vervielfältigung, Bearbeitung, Verbreitung und jede Art der Verwertung außerhalb der Grenzen des Urheberrechtes bedürfen der schriftlichen Zustimmung des jeweiligen Autors bzw. Erstellers.
                    </Typography>

                    <Box sx={{ mt: 6, mb: 2 }}>
                        <Typography variant="h6" gutterBottom>
                            Datenschutzerklärung
                        </Typography>
                        <Divider />
                    </Box>

                    <Typography variant="subtitle1" fontWeight={600} gutterBottom sx={{ mt: 3 }}>
                        1. Datenschutz auf einen Blick
                    </Typography>
                    <Typography variant="body2" fontWeight={600} gutterBottom>
                        Allgemeine Hinweise
                    </Typography>
                    <Typography variant="body2" paragraph>
                        Die folgenden Hinweise geben einen einfachen Überblick darüber, was mit Ihren personenbezogenen Daten passiert, wenn Sie unsere Website besuchen. Personenbezogene Daten sind alle Daten, mit denen Sie persönlich identifiziert werden können.
                    </Typography>

                    <Typography variant="body2" fontWeight={600} gutterBottom>
                        Datenerfassung auf unserer Website
                    </Typography>
                    <Typography variant="body2" paragraph>
                        <strong>Wer ist verantwortlich für die Datenerfassung auf dieser Website?</strong><br />
                        Die Datenverarbeitung auf dieser Website erfolgt durch den Websitebetreiber. Dessen Kontaktdaten können Sie dem Impressum dieser Website entnehmen.
                    </Typography>
                    <Typography variant="body2" paragraph>
                        <strong>Wie erfassen wir Ihre Daten?</strong><br />
                        Ihre Daten werden zum einen dadurch erhoben, dass Sie uns diese mitteilen (z.B. Kontaktformular). Andere Daten werden automatisch beim Besuch der Website durch unsere IT-Systeme erfasst. Das sind vor allem technische Daten (z.B. Internetbrowser, Betriebssystem oder Uhrzeit des Seitenaufrufs).
                    </Typography>
                    <Typography variant="body2" paragraph>
                        <strong>Wofür nutzen wir Ihre Daten?</strong><br />
                        Ein Teil der Daten wird erhoben, um eine fehlerfreie Bereitstellung der Website zu gewährleisten. Andere Daten können zur Analyse Ihres Nutzerverhaltens verwendet werden.
                    </Typography>
                    <Typography variant="body2" paragraph>
                        <strong>Welche Rechte haben Sie bezüglich Ihrer Daten?</strong><br />
                        Sie haben jederzeit das Recht unentgeltlich Auskunft über Herkunft, Empfänger und Zweck Ihrer gespeicherten personenbezogenen Daten zu erhalten. Sie haben außerdem ein Recht, die Berichtigung, Sperrung oder Löschung dieser Daten zu verlangen.
                    </Typography>

                    <Typography variant="subtitle1" fontWeight={600} gutterBottom sx={{ mt: 3 }}>
                        2. Allgemeine Hinweise und Pflichtinformationen
                    </Typography>
                    <Typography variant="body2" paragraph>
                        Die Betreiber dieser Seiten nehmen den Schutz Ihrer persönlichen Daten sehr ernst. Wir behandeln Ihre personenbezogenen Daten vertraulich und entsprechend der gesetzlichen Datenschutzvorschriften sowie dieser Datenschutzerklärung.
                    </Typography>
                    <Typography variant="body2" paragraph>
                        Hinweis zur verantwortlichen Stelle:<br />
                        Fachschaftsvertretung Informatik<br />
                        Westfälische Hochschule<br />
                        Neidenburgerstr. 43<br />
                        45897 Gelsenkirchen<br />
                        E-Mail: kontakt@fachschaftinformatik.de
                    </Typography>
                    <Typography variant="body2" paragraph>
                        <strong>SSL- bzw. TLS-Verschlüsselung</strong><br />
                        Diese Seite nutzt aus Sicherheitsgründen und zum Schutz der Übertragung vertraulicher Inhalte eine SSL-bzw. TLS-Verschlüsselung. Damit können Daten, die Sie an uns übermitteln, nicht von Dritten mitgelesen werden.
                    </Typography>
                    <Typography variant="subtitle1" fontWeight={600} gutterBottom sx={{ mt: 3 }}>
                        3. Datenerfassung auf unserer Website
                    </Typography>
                    <Typography variant="body2" paragraph>
                        <strong>Cookies</strong><br />
                        Die Internetseiten verwenden teilweise so genannte Cookies. Diese dienen dazu, unser Angebot nutzerfreundlicher, effektiver und securer zu machen. Die meisten der von uns verwendeten Cookies sind so genannte “Session-Cookies”, die nach Ende Ihres Besuchs automatisch gelöscht werden.
                    </Typography>
                    <Typography variant="body2" paragraph>
                        <strong>Server-Log-Dateien</strong><br />
                        Der Provider der Seiten erhebt und speichert automatisch Informationen in Server-Log-Dateien, die Ihr Browser automatisch an uns übermittelt (Browsertyp, Betriebssystem, Referrer URL, Hostname, Uhrzeit, IP-Adresse). Eine Zusammenführung dieser Daten mit anderen Datenquellen wird nicht vorgenommen.
                    </Typography>
                </Paper>
            </Box>
        </Sidebar>
    );
}
