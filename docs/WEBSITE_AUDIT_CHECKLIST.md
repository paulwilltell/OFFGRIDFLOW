# OffGridFlow Website Sustainability Audit Checklist

> **Enterprise marketing website audit for performance, accessibility, and sustainability**  
> **Standard**: Google PageSpeed 90+, WCAG 2.1 AA, Sustainable Web Design

---

## 1. Performance Audit

### 1.1 Core Web Vitals

| Metric | Target | Tool | Priority |
|--------|--------|------|----------|
| **LCP** (Largest Contentful Paint) | < 2.5s | Lighthouse | Critical |
| **INP** (Interaction to Next Paint) | < 200ms | Lighthouse | Critical |
| **CLS** (Cumulative Layout Shift) | < 0.1 | Lighthouse | Critical |
| **FCP** (First Contentful Paint) | < 1.8s | Lighthouse | High |
| **TTFB** (Time to First Byte) | < 600ms | WebPageTest | High |

### 1.2 Asset Optimization

| Item | Check | Implementation |
|------|-------|----------------|
| **Image formats** | ☐ Use WebP/AVIF with fallbacks | Next.js Image component |
| **Image sizing** | ☐ Responsive srcset | `sizes` attribute |
| **Image lazy loading** | ☐ Native lazy loading | `loading="lazy"` |
| **Hero images** | ☐ Preload critical images | `<link rel="preload">` |
| **SVG icons** | ☐ Inline critical, sprite others | Icon component |
| **Font loading** | ☐ `font-display: swap` | CSS/Tailwind |
| **Font subsetting** | ☐ Subset to used characters | Build-time |
| **CSS minification** | ☐ Remove unused CSS | PurgeCSS/Tailwind |
| **JS bundling** | ☐ Code splitting by route | Next.js dynamic |
| **JS tree-shaking** | ☐ Dead code elimination | Webpack/Turbopack |

### 1.3 Caching Strategy

| Resource | Cache-Control | Notes |
|----------|---------------|-------|
| Static assets | `max-age=31536000, immutable` | Hash in filename |
| HTML pages | `max-age=0, must-revalidate` | ISR/SSG |
| API responses | `max-age=60, stale-while-revalidate=300` | Per endpoint |
| Fonts | `max-age=31536000` | Static |

### 1.4 Bundle Analysis

```bash
# Commands to analyze bundle
npx @next/bundle-analyzer
npx source-map-explorer build/**/*.js
```

| Check | Target |
|-------|--------|
| Initial JS bundle | < 100KB gzipped |
| First-load JS | < 200KB gzipped |
| Largest chunk | < 150KB gzipped |
| No duplicate packages | Deduplicated |

---

## 2. Accessibility Audit (WCAG 2.1 AA)

### 2.1 Perceivable

| Guideline | Check | Implementation |
|-----------|-------|----------------|
| **1.1.1** Text alternatives | ☐ All images have alt text | `alt` attribute |
| **1.1.1** Decorative images | ☐ Empty alt or `role="presentation"` | `alt=""` |
| **1.3.1** Info and relationships | ☐ Semantic HTML | `<nav>`, `<main>`, `<section>` |
| **1.3.1** Form labels | ☐ Associated labels | `<label htmlFor>` |
| **1.3.2** Meaningful sequence | ☐ Logical DOM order | CSS layout only |
| **1.4.1** Use of color | ☐ Not sole indicator | Icons/text + color |
| **1.4.3** Contrast (minimum) | ☐ 4.5:1 text, 3:1 large | axe DevTools |
| **1.4.4** Resize text | ☐ 200% zoom usable | Responsive design |
| **1.4.10** Reflow | ☐ No horizontal scroll at 320px | Media queries |
| **1.4.11** Non-text contrast | ☐ 3:1 UI components | Border/focus styles |

### 2.2 Operable

| Guideline | Check | Implementation |
|-----------|-------|----------------|
| **2.1.1** Keyboard | ☐ All interactive elements | Native elements |
| **2.1.2** No keyboard trap | ☐ Focus escapes modals | Focus management |
| **2.4.1** Bypass blocks | ☐ Skip to main content | Skip link |
| **2.4.3** Focus order | ☐ Logical tab sequence | DOM order |
| **2.4.4** Link purpose | ☐ Descriptive link text | No "click here" |
| **2.4.6** Headings and labels | ☐ Descriptive headings | `<h1>`-`<h6>` |
| **2.4.7** Focus visible | ☐ Clear focus indicator | `:focus-visible` |
| **2.5.3** Label in name | ☐ Accessible name matches visible | `aria-label` |

### 2.3 Understandable

| Guideline | Check | Implementation |
|-----------|-------|----------------|
| **3.1.1** Language of page | ☐ `lang` attribute | `<html lang="en">` |
| **3.2.1** On focus | ☐ No context change | Standard behavior |
| **3.2.2** On input | ☐ Predictable changes | User-initiated |
| **3.3.1** Error identification | ☐ Error messages | Accessible errors |
| **3.3.2** Labels or instructions | ☐ Input hints | Placeholder + label |

### 2.4 Robust

| Guideline | Check | Implementation |
|-----------|-------|----------------|
| **4.1.1** Parsing | ☐ Valid HTML | W3C validator |
| **4.1.2** Name, Role, Value | ☐ ARIA when needed | Custom components |

### 2.5 Testing Tools

```bash
# Automated testing
npx axe-core-cli https://offgridflow.com
npx pa11y https://offgridflow.com

# Browser extensions
# - axe DevTools
# - WAVE
# - Lighthouse

# Screen reader testing
# - VoiceOver (macOS)
# - NVDA (Windows)
# - Orca (Linux)
```

---

## 3. Sustainability Audit

### 3.1 Carbon Footprint Estimation

| Metric | Target | Tool |
|--------|--------|------|
| Page weight | < 500KB | WebPageTest |
| CO2 per page view | < 0.5g | Website Carbon Calculator |
| Requests per page | < 30 | DevTools Network |
| Third-party requests | < 10 | Request Map |

### 3.2 Sustainable Design Checklist

| Category | Check | Impact |
|----------|-------|--------|
| **Images** | ☐ Compressed, right-sized | High |
| **Video** | ☐ No autoplay, lazy loaded | High |
| **Fonts** | ☐ System fonts or WOFF2 | Medium |
| **Third-parties** | ☐ Minimal, deferred | High |
| **Dark mode** | ☐ OLED-friendly colors | Medium |
| **Caching** | ☐ Aggressive caching | High |
| **Green hosting** | ☐ Renewable energy provider | High |
| **CDN** | ☐ Edge caching | Medium |

### 3.3 Hosting Requirements

| Requirement | Specification |
|-------------|---------------|
| Green certification | Green Web Foundation certified |
| Data center location | EU/US renewable energy grids |
| CDN provider | Cloudflare/Fastly (carbon neutral) |
| Serverless functions | Cold start optimized |

### 3.4 Measurement Tools

```bash
# Website Carbon Calculator
# https://www.websitecarbon.com/

# Beacon (sustainable web)
# https://digitalbeacon.co/

# Green Web Check
# https://www.thegreenwebfoundation.org/green-web-check/
```

---

## 4. SEO Audit

### 4.1 Technical SEO

| Item | Check | Implementation |
|------|-------|----------------|
| **Meta title** | ☐ Unique per page, < 60 chars | `<title>` |
| **Meta description** | ☐ Unique per page, < 160 chars | `<meta name="description">` |
| **Open Graph** | ☐ OG tags for social | `og:title`, `og:image`, etc. |
| **Twitter Card** | ☐ Twitter meta tags | `twitter:card`, etc. |
| **Canonical URL** | ☐ Self-referencing canonical | `<link rel="canonical">` |
| **Structured data** | ☐ JSON-LD schema | Organization, Product |
| **Sitemap** | ☐ XML sitemap | `/sitemap.xml` |
| **Robots.txt** | ☐ Proper directives | `/robots.txt` |
| **404 page** | ☐ Custom, helpful | Next.js 404 page |
| **Redirects** | ☐ 301 for moved pages | `next.config.js` |

### 4.2 Content SEO

| Page | Target Keywords | Priority |
|------|-----------------|----------|
| Homepage | carbon accounting software, ESG platform | High |
| Features | emissions tracking, scope 3 calculator | High |
| Compliance | CSRD compliance software, SEC climate | High |
| Pricing | carbon accounting pricing, ESG software cost | Medium |
| Blog | [various long-tail keywords] | Medium |

### 4.3 Performance SEO

| Metric | Impact | Target |
|--------|--------|--------|
| Core Web Vitals | Ranking factor | All green |
| Mobile-first | Ranking factor | Responsive |
| HTTPS | Ranking factor | Enforced |
| Page speed | User experience | < 3s load |

---

## 5. Security Audit

### 5.1 HTTP Headers

| Header | Value | Purpose |
|--------|-------|---------|
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` | HSTS |
| `X-Content-Type-Options` | `nosniff` | MIME sniffing |
| `X-Frame-Options` | `DENY` | Clickjacking |
| `Content-Security-Policy` | [Strict policy] | XSS prevention |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Privacy |
| `Permissions-Policy` | Restricted | Feature control |

### 5.2 SSL/TLS

| Check | Requirement |
|-------|-------------|
| Certificate validity | Valid, not expiring soon |
| TLS version | 1.2 minimum, prefer 1.3 |
| Certificate chain | Complete chain |
| HSTS preload | Submitted to preload list |

### 5.3 Security Tools

```bash
# Security headers check
# https://securityheaders.com/

# SSL Labs test
# https://www.ssllabs.com/ssltest/

# Mozilla Observatory
# https://observatory.mozilla.org/
```

---

## 6. Implementation Checklist

### 6.1 Next.js Configuration

```typescript
// next.config.js
const nextConfig = {
  // Image optimization
  images: {
    formats: ['image/avif', 'image/webp'],
    deviceSizes: [640, 750, 828, 1080, 1200, 1920],
    minimumCacheTTL: 31536000,
  },
  
  // Headers
  async headers() {
    return [
      {
        source: '/:path*',
        headers: [
          { key: 'X-DNS-Prefetch-Control', value: 'on' },
          { key: 'Strict-Transport-Security', value: 'max-age=31536000; includeSubDomains' },
          { key: 'X-Content-Type-Options', value: 'nosniff' },
          { key: 'X-Frame-Options', value: 'DENY' },
          { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' },
        ],
      },
    ];
  },
  
  // Compression
  compress: true,
  
  // Production optimizations
  productionBrowserSourceMaps: false,
  swcMinify: true,
};
```

### 6.2 Component Checklist

| Component | Accessibility | Performance |
|-----------|---------------|-------------|
| Navigation | ☐ Skip link, ARIA | ☐ Lazy menu items |
| Hero | ☐ Heading hierarchy | ☐ Preload image |
| Features | ☐ Semantic sections | ☐ Lazy load icons |
| Testimonials | ☐ Blockquote, cite | ☐ Lazy load avatars |
| Pricing | ☐ Data tables | ☐ Static render |
| Footer | ☐ Nav landmarks | ☐ Minimal JS |
| Forms | ☐ Labels, errors | ☐ Debounced validation |

### 6.3 Pre-Launch Checklist

```markdown
## Pre-Launch

- [ ] Lighthouse score 90+ (all categories)
- [ ] axe scan: 0 critical/serious issues
- [ ] Mobile responsive: all breakpoints
- [ ] Dark mode: functional
- [ ] Forms: accessible errors
- [ ] Images: all have alt text
- [ ] Fonts: WOFF2, subset
- [ ] Bundle: < 200KB initial
- [ ] Security headers: all set
- [ ] SSL: A+ rating
- [ ] Sitemap: generated
- [ ] robots.txt: configured
- [ ] Analytics: privacy-respecting
- [ ] Error pages: 404, 500
- [ ] Green hosting: confirmed
```

---

## 7. Monitoring

### 7.1 Ongoing Monitoring

| Tool | Metric | Frequency |
|------|--------|-----------|
| Lighthouse CI | Performance score | Every deploy |
| SpeedCurve/Calibre | Core Web Vitals | Daily |
| axe-monitor | Accessibility | Weekly |
| Sentry | Errors | Real-time |
| Google Search Console | SEO issues | Weekly |

### 7.2 Alerting Thresholds

| Metric | Warning | Critical |
|--------|---------|----------|
| LCP | > 2.5s | > 4s |
| INP | > 200ms | > 500ms |
| CLS | > 0.1 | > 0.25 |
| Lighthouse perf | < 90 | < 70 |
| Error rate | > 0.1% | > 1% |

---

## Summary

This audit checklist ensures the OffGridFlow marketing website meets enterprise standards for:

- **Performance**: 90+ Lighthouse score, optimized Core Web Vitals
- **Accessibility**: WCAG 2.1 AA compliance
- **Sustainability**: Low carbon footprint, green hosting
- **SEO**: Optimized for search visibility
- **Security**: Hardened headers, TLS 1.3

**Recommended implementation order:**
1. Performance optimizations (highest user impact)
2. Accessibility fixes (legal/ethical requirement)
3. Security headers (quick wins)
4. SEO enhancements (ongoing)
5. Sustainability improvements (brand alignment)

---

**End of Website Audit Checklist**
