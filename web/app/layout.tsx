import "../styles/globals.css";
import type { ReactNode } from "react";
import { AppProviders } from "./providers";
import { AppHeader } from "./components/AppHeader";

export const metadata = {
  title: "OffGridFlow",
  description: "Carbon accounting and compliance platform",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body>
        <AppProviders>
          <AppHeader />
          <main style={{ padding: "1.5rem", minHeight: "100vh" }}>{children}</main>
          <footer
            style={{
              padding: "1rem 1.5rem",
              borderTop: "1px solid #1d2940",
              color: "#666",
              fontSize: "0.8rem",
              textAlign: "center",
            }}
          >
            © {new Date().getFullYear()} OffGridFlow - Carbon Accounting Platform
          </footer>
        </AppProviders>
      </body>
    </html>
  );
}
