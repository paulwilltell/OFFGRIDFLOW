import { NextRequest, NextResponse } from 'next/server';

export async function POST(req: NextRequest) {
  try {
    const body = await req.json();

    const {
      from,
      contact_name,
      company_name,
      company_size,
      industry,
      number_of_sites,
      compliance_frameworks,
      regions_of_operation,
      implementation_timeline,
      current_plan,
      additional_notes,
    } = body;

    if (!from || !contact_name || !company_name) {
      return NextResponse.json({ error: 'Missing required fields' }, { status: 400 });
    }

    const emailBody = `
NEW CARBON COMMAND ELITE INQUIRY
=================================

FROM:       ${contact_name} <${from}>
COMPANY:    ${company_name}
SIZE:       ${company_size || 'Not specified'}
INDUSTRY:   ${industry || 'Not specified'}
SITES:      ${number_of_sites}
REGIONS:    ${regions_of_operation || 'Not specified'}
TIMELINE:   ${implementation_timeline || 'Not specified'}
FRAMEWORKS: ${compliance_frameworks?.length ? compliance_frameworks.join(', ') : 'None selected'}
CURRENT PLAN: ${current_plan || 'Free / Not subscribed'}

ADDITIONAL NOTES:
${additional_notes || 'None provided'}

=================================
Reply directly to: ${from}
    `.trim();

    const sgRes = await fetch('https://api.sendgrid.com/v3/mail/send', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${process.env.SENDGRID_API_KEY}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        personalizations: [{
          to: [{ email: 'paul@offgridflow.com', name: 'Paul — OffGridFlow' }],
          reply_to: { email: from, name: contact_name },
        }],
        from: {
          email: process.env.SMTP_FROM_EMAIL ?? 'noreply@offgridflow.com',
          name: 'OffGridFlow Enterprise Inquiries',
        },
        subject: `Elite Inquiry: ${company_name} — ${number_of_sites} site${number_of_sites > 1 ? 's' : ''} — ${implementation_timeline || 'Timeline TBD'}`,
        content: [
          { type: 'text/plain', value: emailBody },
          {
            type: 'text/html',
            value: `<!DOCTYPE html>
<html>
<body style="font-family:system-ui,sans-serif;color:#111;max-width:600px;margin:0 auto;padding:32px;">
  <div style="background:#064e3b;color:white;padding:24px 32px;border-radius:12px 12px 0 0;">
    <p style="margin:0;font-size:11px;letter-spacing:2px;text-transform:uppercase;color:#6ee7b7;">Carbon Command Elite</p>
    <h1 style="margin:8px 0 0;font-size:22px;">New Enterprise Inquiry</h1>
  </div>
  <div style="background:#f9fafb;border:1px solid #e5e7eb;border-top:none;padding:24px 32px;border-radius:0 0 12px 12px;">
    <table style="width:100%;border-collapse:collapse;font-size:14px;">
      <tr><td style="padding:8px 0;color:#6b7280;width:160px;">Contact</td><td style="padding:8px 0;font-weight:600;">${contact_name}</td></tr>
      <tr><td style="padding:8px 0;color:#6b7280;">Email</td><td style="padding:8px 0;"><a href="mailto:${from}" style="color:#059669;">${from}</a></td></tr>
      <tr><td style="padding:8px 0;color:#6b7280;">Company</td><td style="padding:8px 0;font-weight:600;">${company_name}</td></tr>
      <tr><td style="padding:8px 0;color:#6b7280;">Size</td><td style="padding:8px 0;">${company_size || '—'}</td></tr>
      <tr><td style="padding:8px 0;color:#6b7280;">Industry</td><td style="padding:8px 0;">${industry || '—'}</td></tr>
      <tr><td style="padding:8px 0;color:#6b7280;">Sites / Facilities</td><td style="padding:8px 0;font-weight:600;">${number_of_sites}</td></tr>
      <tr><td style="padding:8px 0;color:#6b7280;">Regions</td><td style="padding:8px 0;">${regions_of_operation || '—'}</td></tr>
      <tr><td style="padding:8px 0;color:#6b7280;">Timeline</td><td style="padding:8px 0;">${implementation_timeline || '—'}</td></tr>
      <tr><td style="padding:8px 0;color:#6b7280;">Frameworks</td><td style="padding:8px 0;">${compliance_frameworks?.length ? compliance_frameworks.join(', ') : '—'}</td></tr>
      <tr><td style="padding:8px 0;color:#6b7280;">Current Plan</td><td style="padding:8px 0;">${current_plan || 'Free / Not subscribed'}</td></tr>
    </table>
    ${additional_notes ? `<div style="margin-top:20px;padding:16px;background:white;border:1px solid #e5e7eb;border-radius:8px;"><p style="margin:0 0 8px;font-size:12px;color:#6b7280;text-transform:uppercase;letter-spacing:1px;">Additional Notes</p><p style="margin:0;font-size:14px;line-height:1.6;">${additional_notes}</p></div>` : ''}
    <div style="margin-top:24px;padding:16px;background:#ecfdf5;border-radius:8px;font-size:13px;color:#065f46;">
      Reply directly to this email to respond to ${contact_name}.
    </div>
  </div>
</body>
</html>`,
          },
        ],
      }),
    });

    if (!sgRes.ok) {
      const errText = await sgRes.text();
      console.error('SendGrid error:', errText);
      return NextResponse.json({ error: 'Email delivery failed' }, { status: 500 });
    }

    return NextResponse.json({ success: true });
  } catch (err) {
    console.error('Elite inquiry error:', err);
    return NextResponse.json({ error: 'Internal error' }, { status: 500 });
  }
}
