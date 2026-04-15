import "../styles/globals.css";
import type { ReactNode } from "react";
import { Inter, JetBrains_Mono } from "next/font/google";
import { AppProviders } from "./providers";
import { CookieConsent } from "./components/CookieConsent";
import { JsonLd, organizationSchema, softwareApplicationSchema } from "@/components/JsonLd";

const inter = Inter({
  subsets: ["latin"],
  display: "swap",
  variable: "--font-sans",
});

const jetBrainsMono = JetBrains_Mono({
  subsets: ["latin"],
  display: "swap",
  variable: "--font-mono",
});

export const metadata = {
  title: "OffGridFlow",
  description: "Carbon accounting and compliance platform",
  icons: {
    icon: [
      { url: '/favicon.ico', sizes: '32x32' },
      { url: '/icon.svg', type: 'image/svg+xml' },
    ],
  },
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html
      lang="en"
      className={`${inter.variable} ${jetBrainsMono.variable}`}
    >
      <head>
        {/* Google Consent Mode v2 defaults — deny all consented storage until
            the user explicitly accepts via the CookieConsent banner. This makes
            the gtag below GDPR/ePrivacy compliant by default. */}
        <script
          dangerouslySetInnerHTML={{
            __html: `
              window.dataLayer = window.dataLayer || [];
              function gtag(){dataLayer.push(arguments);}
              gtag('consent', 'default', {
                'ad_storage': 'denied',
                'ad_user_data': 'denied',
                'ad_personalization': 'denied',
                'analytics_storage': 'denied',
                'wait_for_update': 500
              });
              gtag('js', new Date());
            `,
          }}
        />
        <script async src="https://www.googletagmanager.com/gtag/js?id=AW-18088629744" />
        <script
          dangerouslySetInnerHTML={{
            __html: `
              gtag('config', 'AW-18088629744', { anonymize_ip: true });
            `,
          }}
        />
      </head>
      <body className="font-sans antialiased">
        {/* Site-wide entity markup. Both schemas appear on every page so that
            search engines consistently associate OffGridFlow LLC as the
            publisher and the platform as the SoftwareApplication product. */}
        <JsonLd id="ld-global-organization" data={organizationSchema()} />
        <JsonLd id="ld-global-software" data={softwareApplicationSchema()} />
        <AppProviders>{children}</AppProviders>
        <CookieConsent />
      </body>
    </html>
  );
}
