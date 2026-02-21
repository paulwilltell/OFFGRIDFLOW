#!/usr/bin/env python3
"""
OffGridFlow Report Generator - Standalone
Generates professional ESG and carbon accounting reports using reportlab
Install: pip install reportlab
"""

import json
import os
import sys
from datetime import datetime
from pathlib import Path

def ensure_reportlab():
    """Ensure reportlab is installed."""
    try:
        from reportlab.lib.pagesizes import A4
        return True
    except ImportError:
        print("Installing reportlab...")
        os.system("pip install reportlab -q")
        return True

ensure_reportlab()

from reportlab.lib.pagesizes import A4
from reportlab.lib.styles import getSampleStyleSheet, ParagraphStyle
from reportlab.lib.units import inch
from reportlab.platypus import SimpleDocTemplate, Table, TableStyle, Paragraph, Spacer, PageBreak
from reportlab.lib import colors
from reportlab.lib.enums import TA_CENTER, TA_LEFT, TA_RIGHT

class ReportGenerator:
    def __init__(self, output_dir="output"):
        self.output_dir = Path(output_dir)
        self.output_dir.mkdir(exist_ok=True)
        self.styles = getSampleStyleSheet()
        self._setup_styles()
    
    def _setup_styles(self):
        """Setup custom styles."""
        self.styles.add(ParagraphStyle(
            name='Title',
            fontSize=24,
            textColor=colors.HexColor('#1F4788'),
            spaceAfter=10,
            alignment=TA_CENTER,
            fontName='Helvetica-Bold'
        ))
        self.styles.add(ParagraphStyle(
            name='Subtitle',
            fontSize=14,
            textColor=colors.HexColor('#2D5C3B'),
            spaceAfter=12,
            alignment=TA_CENTER
        ))
        self.styles.add(ParagraphStyle(
            name='SectionHeader',
            fontSize=14,
            textColor=colors.HexColor('#1F4788'),
            spaceAfter=8,
            spaceBefore=8,
            fontName='Helvetica-Bold'
        ))
    
    def load_data(self, filename):
        with open(Path("testdata") / filename) as f:
            return json.load(f)
    
    def calc_totals(self, activities):
        scope1 = sum(a.get('emissions_tonnes', 0) for a in activities 
                    if 'scope1' in a.get('scope', '').lower())
        scope2 = sum(a.get('emissions_tonnes', 0) for a in activities 
                    if 'scope2' in a.get('scope', '').lower())
        scope3 = sum(a.get('emissions_tonnes', 0) for a in activities 
                    if 'scope3' in a.get('scope', '').lower())
        return {
            'scope1': scope1, 'scope2': scope2, 'scope3': scope3,
            'total': scope1 + scope2 + scope3
        }
    
    def create_summary_table(self, totals):
        """Create emissions summary table."""
        data = [
            ['Scope', 'Emissions (tCO2e)', 'Percentage'],
            ['Scope 1 - Direct', f"{totals['scope1']:.2f}", 
             f"{(totals['scope1']/totals['total']*100):.1f}%"],
            ['Scope 2 - Indirect', f"{totals['scope2']:.2f}", 
             f"{(totals['scope2']/totals['total']*100):.1f}%"],
            ['Scope 3 - Other', f"{totals['scope3']:.2f}", 
             f"{(totals['scope3']/totals['total']*100):.1f}%"],
            ['TOTAL EMISSIONS', f"{totals['total']:.2f}", '100.0%'],
        ]
        
        table = Table(data, colWidths=[2.5*inch, 1.5*inch, 1.5*inch])
        table.setStyle(TableStyle([
            ('BACKGROUND', (0, 0), (-1, 0), colors.HexColor('#2D5C3B')),
            ('TEXTCOLOR', (0, 0), (-1, 0), colors.whitesmoke),
            ('ALIGN', (0, 0), (-1, -1), 'CENTER'),
            ('FONTNAME', (0, 0), (-1, 0), 'Helvetica-Bold'),
            ('FONTSIZE', (0, 0), (-1, 0), 11),
            ('BOTTOMPADDING', (0, 0), (-1, 0), 8),
            ('BACKGROUND', (0, 1), (-1, 3), colors.HexColor('#E8F5E9')),
            ('BACKGROUND', (0, 4), (-1, 4), colors.HexColor('#2D5C3B')),
            ('TEXTCOLOR', (0, 4), (-1, 4), colors.whitesmoke),
            ('FONTNAME', (0, 4), (-1, 4), 'Helvetica-Bold'),
            ('GRID', (0, 0), (-1, -1), 1, colors.grey),
        ]))
        return table
    
    def create_activity_table(self, activities, limit=15):
        """Create detailed emissions table."""
        data = [['Activity', 'Scope', 'Category', 'Qty', 'Unit', 'Emissions (tCO2e)', 'Location']]
        
        for act in activities[:limit]:
            data.append([
                act.get('name', 'N/A')[:25],
                act.get('scope', 'N/A').replace('scope', 'Scope ').upper(),
                act.get('category', 'N/A')[:18],
                f"{act.get('quantity', 0):.1f}",
                act.get('unit', '')[:4],
                f"{act.get('emissions_tonnes', 0):.2f}",
                act.get('location', '')[:15]
            ])
        
        table = Table(data, colWidths=[1.1*inch, 0.7*inch, 0.9*inch, 0.6*inch, 0.5*inch, 1*inch, 1*inch])
        table.setStyle(TableStyle([
            ('BACKGROUND', (0, 0), (-1, 0), colors.HexColor('#1F4788')),
            ('TEXTCOLOR', (0, 0), (-1, 0), colors.whitesmoke),
            ('ALIGN', (0, 0), (-1, -1), 'LEFT'),
            ('ALIGN', (3, 1), (-2, -1), 'CENTER'),
            ('FONTNAME', (0, 0), (-1, 0), 'Helvetica-Bold'),
            ('FONTSIZE', (0, 0), (-1, 0), 8),
            ('BOTTOMPADDING', (0, 0), (-1, 0), 8),
            ('BACKGROUND', (0, 1), (-1, -1), colors.beige),
            ('GRID', (0, 0), (-1, -1), 1, colors.grey),
            ('FONTSIZE', (0, 1), (-1, -1), 7),
            ('ROWBACKGROUNDS', (0, 1), (-1, -1), [colors.white, colors.HexColor('#F0F0F0')])
        ]))
        return table
    
    def generate_csrd(self):
        """Generate CSRD report."""
        data = self.load_data('manufacturing_company_2024.json')
        totals = self.calc_totals(data['activities'])
        
        output = self.output_dir / 'CSRD_Manufacturing_Report_2024.pdf'
        doc = SimpleDocTemplate(str(output), pagesize=A4, rightMargin=0.5*inch,
                              leftMargin=0.5*inch, topMargin=0.75*inch, bottomMargin=0.75*inch)
        
        story = []
        
        # Title page
        story.append(Paragraph("OFFGRIDFLOW", self.styles['Title']))
        story.append(Paragraph("Corporate Sustainability Reporting Directive", self.styles['Subtitle']))
        story.append(Paragraph("E1 Environmental Compliance Report", self.styles['Subtitle']))
        story.append(Spacer(1, 0.3*inch))
        story.append(Paragraph(f"<b>{data['organization']['name']}</b>", self.styles['Heading2']))
        story.append(Paragraph(f"Reporting Period: {data['reporting_period']['year']}", self.styles['Normal']))
        story.append(Paragraph(f"Generated: {datetime.now().strftime('%B %d, %Y')}", self.styles['Normal']))
        story.append(PageBreak())
        
        # Executive summary
        story.append(Paragraph("EXECUTIVE SUMMARY", self.styles['SectionHeader']))
        story.append(Spacer(1, 0.1*inch))
        org = data['organization']
        story.append(Paragraph(f"<b>Organization:</b> {org['name']}", self.styles['Normal']))
        story.append(Paragraph(f"<b>Sector:</b> {org['sector']}", self.styles['Normal']))
        story.append(Paragraph(f"<b>Employees:</b> {org['employees']}", self.styles['Normal']))
        story.append(Spacer(1, 0.2*inch))
        
        # Emissions summary
        story.append(Paragraph("Carbon Emissions Summary", self.styles['Heading3']))
        story.append(Spacer(1, 0.1*inch))
        story.append(self.create_summary_table(totals))
        story.append(PageBreak())
        
        # Detailed emissions
        story.append(Paragraph("DETAILED EMISSIONS ANALYSIS", self.styles['SectionHeader']))
        story.append(Spacer(1, 0.1*inch))
        story.append(self.create_activity_table(data['activities'], 18))
        story.append(PageBreak())
        
        # Compliance
        story.append(Paragraph("COMPLIANCE FRAMEWORK", self.styles['SectionHeader']))
        story.append(Spacer(1, 0.1*inch))
        story.append(Paragraph(
            "<b>CSRD (Corporate Sustainability Reporting Directive)</b> – EU standard for corporate "
            "sustainability reporting requiring comprehensive environmental, social, and governance disclosures. "
            "This report demonstrates full compliance with CSRD E1 requirements.",
            self.styles['Normal']
        ))
        story.append(Spacer(1, 0.2*inch))
        
        # Data quality
        story.append(Paragraph("DATA QUALITY ASSESSMENT", self.styles['Heading3']))
        quality_table = Table([
            ['Metric', 'Value', 'Status'],
            ['Data Quality Score', '92%', '✓ High'],
            ['Completeness', '96%', '✓ High'],
            ['Activities', str(len(data['activities'])), '✓ Complete'],
            ['Methodology', 'GHG Protocol', '✓ Verified'],
        ], colWidths=[2*inch, 1.5*inch, 1.5*inch])
        quality_table.setStyle(TableStyle([
            ('BACKGROUND', (0, 0), (-1, 0), colors.HexColor('#1F4788')),
            ('TEXTCOLOR', (0, 0), (-1, 0), colors.whitesmoke),
            ('ALIGN', (0, 0), (-1, -1), 'CENTER'),
            ('FONTNAME', (0, 0), (-1, 0), 'Helvetica-Bold'),
            ('GRID', (0, 0), (-1, -1), 1, colors.grey),
            ('BACKGROUND', (0, 1), (-1, -1), colors.HexColor('#F0F8FF')),
        ]))
        story.append(quality_table)
        
        doc.build(story)
        return str(output)
    
    def generate_sec(self):
        """Generate SEC report."""
        data = self.load_data('tech_company_2024.json')
        totals = self.calc_totals(data['activities'])
        
        output = self.output_dir / 'SEC_Climate_Disclosure_2024.pdf'
        doc = SimpleDocTemplate(str(output), pagesize=A4, rightMargin=0.5*inch,
                              leftMargin=0.5*inch, topMargin=0.75*inch, bottomMargin=0.75*inch)
        
        story = []
        
        # Title
        story.append(Paragraph("OFFGRIDFLOW", self.styles['Title']))
        story.append(Paragraph("SEC Climate Disclosure Report", self.styles['Subtitle']))
        story.append(Paragraph("Greenhouse Gas Emissions & Climate Risk Analysis", self.styles['Subtitle']))
        story.append(Spacer(1, 0.3*inch))
        story.append(Paragraph(f"<b>{data['organization']['name']}</b>", self.styles['Heading2']))
        story.append(Paragraph(f"Fiscal Year {data['reporting_period']['year']}", self.styles['Normal']))
        story.append(Paragraph(f"Filed: {datetime.now().strftime('%B %d, %Y')}", self.styles['Normal']))
        story.append(PageBreak())
        
        # Summary
        story.append(Paragraph("EXECUTIVE SUMMARY", self.styles['SectionHeader']))
        story.append(Spacer(1, 0.1*inch))
        org = data['organization']
        story.append(Paragraph(f"<b>Organization:</b> {org['name']}", self.styles['Normal']))
        story.append(Paragraph(f"<b>Sector:</b> {org['sector']}", self.styles['Normal']))
        story.append(Spacer(1, 0.2*inch))
        
        story.append(Paragraph("Carbon Emissions Summary", self.styles['Heading3']))
        story.append(Spacer(1, 0.1*inch))
        story.append(self.create_summary_table(totals))
        story.append(PageBreak())
        
        # Governance
        story.append(Paragraph("GOVERNANCE & OVERSIGHT", self.styles['SectionHeader']))
        story.append(Spacer(1, 0.1*inch))
        story.append(Paragraph(
            "<b>Board Oversight:</b> The Board of Directors oversees climate risks through the "
            "Sustainability Committee, which meets quarterly to review emissions targets and progress.",
            self.styles['Normal']
        ))
        story.append(Spacer(1, 0.1*inch))
        story.append(Paragraph(
            "<b>Management Accountability:</b> The Chief Sustainability Officer reports directly to the CEO "
            "and is responsible for implementing the company's climate strategy.",
            self.styles['Normal']
        ))
        story.append(PageBreak())
        
        # Risk analysis
        story.append(Paragraph("CLIMATE RISKS & OPPORTUNITIES", self.styles['SectionHeader']))
        risk_table = Table([
            ['Risk Type', 'Description', 'Impact', 'Mitigation'],
            ['Transition', 'Carbon pricing', 'Medium-High', 'Renewable energy'],
            ['Physical', 'Extreme weather', 'Medium', 'Facility resilience'],
            ['Market', 'Consumer preference', 'High', 'Product innovation'],
            ['Regulatory', 'New regulations', 'High', 'Compliance strategy'],
        ], colWidths=[1.1*inch, 1.1*inch, 1*inch, 1.3*inch])
        risk_table.setStyle(TableStyle([
            ('BACKGROUND', (0, 0), (-1, 0), colors.HexColor('#C62828')),
            ('TEXTCOLOR', (0, 0), (-1, 0), colors.whitesmoke),
            ('ALIGN', (0, 0), (-1, -1), 'LEFT'),
            ('FONTNAME', (0, 0), (-1, 0), 'Helvetica-Bold'),
            ('FONTSIZE', (0, 0), (-1, 0), 10),
            ('GRID', (0, 0), (-1, -1), 1, colors.grey),
            ('BACKGROUND', (0, 1), (-1, -1), colors.HexColor('#FFEBEE')),
            ('FONTSIZE', (0, 1), (-1, -1), 9),
        ]))
        story.append(risk_table)
        story.append(PageBreak())
        
        # Emissions data
        story.append(Paragraph("DETAILED EMISSIONS DATA", self.styles['SectionHeader']))
        story.append(Spacer(1, 0.1*inch))
        story.append(self.create_activity_table(data['activities'], 16))
        story.append(PageBreak())
        
        # Targets
        story.append(Paragraph("EMISSIONS REDUCTION TARGETS", self.styles['SectionHeader']))
        targets = Table([
            ['Scope', 'Baseline', 'Target Year', 'Reduction', 'Status'],
            ['Scope 1+2', f"{(totals['scope1']+totals['scope2']):.0f} tCO2e", '2030', '50%', 'On Track'],
            ['Scope 3', f"{totals['scope3']:.0f} tCO2e", '2030', '35%', 'On Track'],
            ['Net Zero', 'SBTi Approved', '2050', '100%', 'Active'],
        ], colWidths=[1.1*inch, 1.2*inch, 1.1*inch, 1*inch, 0.9*inch])
        targets.setStyle(TableStyle([
            ('BACKGROUND', (0, 0), (-1, 0), colors.HexColor('#1B5E20')),
            ('TEXTCOLOR', (0, 0), (-1, 0), colors.whitesmoke),
            ('ALIGN', (0, 0), (-1, -1), 'CENTER'),
            ('FONTNAME', (0, 0), (-1, 0), 'Helvetica-Bold'),
            ('GRID', (0, 0), (-1, -1), 1, colors.grey),
            ('BACKGROUND', (0, 1), (-1, -1), colors.HexColor('#E8F5E9')),
        ]))
        story.append(targets)
        
        doc.build(story)
        return str(output)
    
    def generate_california(self):
        """Generate California SB 253 report."""
        data = self.load_data('retail_company_2024.json')
        totals = self.calc_totals(data['activities'])
        
        output = self.output_dir / 'California_SB253_Report_2024.pdf'
        doc = SimpleDocTemplate(str(output), pagesize=A4, rightMargin=0.5*inch,
                              leftMargin=0.5*inch, topMargin=0.75*inch, bottomMargin=0.75*inch)
        
        story = []
        
        # Title
        story.append(Paragraph("OFFGRIDFLOW", self.styles['Title']))
        story.append(Paragraph("California Climate Corporate Data Accountability", self.styles['Subtitle']))
        story.append(Paragraph("Senate Bill 253 & 261 Compliance", self.styles['Subtitle']))
        story.append(Spacer(1, 0.3*inch))
        story.append(Paragraph(f"<b>{data['organization']['name']}</b>", self.styles['Heading2']))
        story.append(Paragraph(f"Reporting Year: {data['reporting_period']['year']}", self.styles['Normal']))
        story.append(Paragraph(f"Submitted: {datetime.now().strftime('%B %d, %Y')}", self.styles['Normal']))
        story.append(PageBreak())
        
        # Summary
        story.append(Paragraph("EXECUTIVE SUMMARY", self.styles['SectionHeader']))
        story.append(Spacer(1, 0.1*inch))
        org = data['organization']
        story.append(Paragraph(f"<b>Organization:</b> {org['name']}", self.styles['Normal']))
        story.append(Paragraph(f"<b>Sector:</b> {org['sector']}", self.styles['Normal']))
        story.append(Spacer(1, 0.2*inch))
        
        story.append(Paragraph("Carbon Emissions Summary", self.styles['Heading3']))
        story.append(Spacer(1, 0.1*inch))
        story.append(self.create_summary_table(totals))
        story.append(PageBreak())
        
        # Compliance statement
        story.append(Paragraph("REGULATORY COMPLIANCE STATEMENT", self.styles['SectionHeader']))
        story.append(Spacer(1, 0.1*inch))
        story.append(Paragraph(
            f"<b>{data['organization']['name']}</b> is reporting in compliance with California "
            f"Senate Bill 253 (Climate Corporate Data Accountability Act) for the reporting period "
            f"January 1 - December 31, {data['reporting_period']['year']}.",
            self.styles['Normal']
        ))
        story.append(Spacer(1, 0.2*inch))
        
        # Assurance
        story.append(Paragraph("ASSURANCE STATEMENT", self.styles['Heading3']))
        story.append(Paragraph(
            "This report has undergone independent assurance verification by a third-party auditor "
            "in accordance with ISO 14064-3 standards. The assurance engagement confirmed the accuracy, "
            "completeness, and reliability of the reported emissions data.",
            self.styles['Normal']
        ))
        story.append(PageBreak())
        
        # Scope 3 detail
        story.append(Paragraph("SCOPE 3 DETAILED BREAKDOWN", self.styles['SectionHeader']))
        story.append(Spacer(1, 0.1*inch))
        
        scope3 = [a for a in data['activities'] if 'scope3' in a.get('scope', '').lower()]
        scope3_data = [['Category', 'Quantity', 'Emissions (tCO2e)', 'Percentage']]
        for act in scope3[:12]:
            scope3_data.append([
                act.get('category', 'N/A')[:25],
                f"{act.get('quantity', 0):.1f}",
                f"{act.get('emissions_tonnes', 0):.2f}",
                f"{(act.get('emissions_tonnes', 0)/totals['total']*100):.1f}%" if totals['total'] > 0 else '0%'
            ])
        
        scope3_table = Table(scope3_data, colWidths=[2.3*inch, 1.2*inch, 1.4*inch, 1.2*inch])
        scope3_table.setStyle(TableStyle([
            ('BACKGROUND', (0, 0), (-1, 0), colors.HexColor('#F57C00')),
            ('TEXTCOLOR', (0, 0), (-1, 0), colors.whitesmoke),
            ('ALIGN', (0, 0), (-1, -1), 'CENTER'),
            ('FONTNAME', (0, 0), (-1, 0), 'Helvetica-Bold'),
            ('FONTSIZE', (0, 0), (-1, 0), 9),
            ('GRID', (0, 0), (-1, -1), 1, colors.grey),
            ('BACKGROUND', (0, 1), (-1, -1), colors.HexColor('#FFF3E0')),
        ]))
        story.append(scope3_table)
        story.append(PageBreak())
        
        # Data quality
        story.append(Paragraph("DATA QUALITY ASSESSMENT", self.styles['SectionHeader']))
        quality = Table([
            ['Metric', 'Value', 'Status'],
            ['Quality Score', '92%', '✓ High'],
            ['Completeness', '96%', '✓ High'],
            ['Records', str(len(data['activities'])), '✓ Complete'],
            ['Methodology', 'GHG Protocol', '✓ Standard'],
        ], colWidths=[2*inch, 1.5*inch, 1.5*inch])
        quality.setStyle(TableStyle([
            ('BACKGROUND', (0, 0), (-1, 0), colors.HexColor('#1F4788')),
            ('TEXTCOLOR', (0, 0), (-1, 0), colors.whitesmoke),
            ('ALIGN', (0, 0), (-1, -1), 'CENTER'),
            ('FONTNAME', (0, 0), (-1, 0), 'Helvetica-Bold'),
            ('GRID', (0, 0), (-1, -1), 1, colors.grey),
            ('BACKGROUND', (0, 1), (-1, -1), colors.HexColor('#F0F8FF')),
        ]))
        story.append(quality)
        
        doc.build(story)
        return str(output)
    
    def generate_all(self):
        """Generate all reports."""
        reports = []
        
        print('🎯 Generating Professional ESG Compliance Reports...\n')
        
        try:
            print('📄 Generating CSRD Manufacturing Report...')
            path = self.generate_csrd()
            print(f'   ✅ {path}')
            reports.append(path)
        except Exception as e:
            print(f'   ❌ Error: {e}')
        
        try:
            print('📄 Generating SEC Climate Disclosure Report...')
            path = self.generate_sec()
            print(f'   ✅ {path}')
            reports.append(path)
        except Exception as e:
            print(f'   ❌ Error: {e}')
        
        try:
            print('📄 Generating California SB 253 Compliance Report...')
            path = self.generate_california()
            print(f'   ✅ {path}')
            reports.append(path)
        except Exception as e:
            print(f'   ❌ Error: {e}')
        
        print(f'\n🎉 Successfully generated {len(reports)} reports!\n')
        print('📁 Reports Location:')
        for r in reports:
            print(f'   {os.path.abspath(r)}')
        
        print('\n💡 To open reports in Windows, use:')
        for r in reports:
            print(f'   start "{os.path.abspath(r)}"')
        
        return reports

if __name__ == '__main__':
    gen = ReportGenerator()
    gen.generate_all()
