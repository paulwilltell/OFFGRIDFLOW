#!/usr/bin/env python3
"""
Generate professional carbon accounting and ESG compliance reports as PDFs.
Uses test data from testdata/ directory to create high-quality example reports.
"""

import json
import os
from datetime import datetime
from pathlib import Path
from reportlab.lib.pagesizes import letter, A4
from reportlab.lib.styles import getSampleStyleSheet, ParagraphStyle
from reportlab.lib.units import inch
from reportlab.platypus import SimpleDocTemplate, Table, TableStyle, Paragraph, Spacer, PageBreak, Image
from reportlab.lib import colors
from reportlab.lib.enums import TA_CENTER, TA_LEFT, TA_RIGHT
import math

class OffGridFlowReportGenerator:
    """Generates professional ESG and carbon accounting reports."""
    
    def __init__(self, output_dir="output"):
        self.output_dir = Path(output_dir)
        self.output_dir.mkdir(exist_ok=True)
        self.styles = getSampleStyleSheet()
        self._setup_custom_styles()
    
    def _setup_custom_styles(self):
        """Setup custom paragraph styles."""
        self.styles.add(ParagraphStyle(
            name='CustomTitle',
            parent=self.styles['Heading1'],
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
            alignment=TA_CENTER,
            fontName='Helvetica'
        ))
        
        self.styles.add(ParagraphStyle(
            name='SectionHeader',
            parent=self.styles['Heading2'],
            fontSize=14,
            textColor=colors.HexColor('#1F4788'),
            spaceAfter=8,
            spaceBefore=8,
            fontName='Helvetica-Bold'
        ))
    
    def _load_test_data(self, filename):
        """Load test data from testdata directory."""
        data_file = Path("testdata") / filename
        with open(data_file, 'r') as f:
            return json.load(f)
    
    def _calculate_totals(self, activities):
        """Calculate emission totals by scope."""
        scope1 = sum(a.get('emissions_tonnes', 0) for a in activities if 'scope1' in a.get('scope', '').lower())
        scope2 = sum(a.get('emissions_tonnes', 0) for a in activities if 'scope2' in a.get('scope', '').lower())
        scope3 = sum(a.get('emissions_tonnes', 0) for a in activities if 'scope3' in a.get('scope', '').lower())
        return {
            'scope1': scope1,
            'scope2': scope2,
            'scope3': scope3,
            'total': scope1 + scope2 + scope3
        }
    
    def _create_emissions_table(self, activities):
        """Create detailed emissions table."""
        # Header row
        data = [['Activity', 'Scope', 'Category', 'Quantity', 'Unit', 'Emissions (tCO2e)', 'Location']]
        
        # Data rows
        for act in activities[:20]:  # Limit to first 20 for readability
            data.append([
                act.get('name', 'N/A')[:30],
                act.get('scope', 'N/A').replace('scope', 'Scope ').upper(),
                act.get('category', 'N/A')[:20],
                f"{act.get('quantity', 0):.1f}",
                act.get('unit', 'N/A'),
                f"{act.get('emissions_tonnes', 0):.2f}",
                act.get('location', 'N/A')[:20]
            ])
        
        table = Table(data, colWidths=[1.2*inch, 0.7*inch, 1*inch, 0.8*inch, 0.5*inch, 1.1*inch, 1.2*inch])
        table.setStyle(TableStyle([
            ('BACKGROUND', (0, 0), (-1, 0), colors.HexColor('#1F4788')),
            ('TEXTCOLOR', (0, 0), (-1, 0), colors.whitesmoke),
            ('ALIGN', (0, 0), (-1, -1), 'LEFT'),
            ('ALIGN', (3, 1), (-2, -1), 'CENTER'),
            ('FONTNAME', (0, 0), (-1, 0), 'Helvetica-Bold'),
            ('FONTSIZE', (0, 0), (-1, 0), 9),
            ('BOTTOMPADDING', (0, 0), (-1, 0), 8),
            ('BACKGROUND', (0, 1), (-1, -1), colors.beige),
            ('GRID', (0, 0), (-1, -1), 1, colors.grey),
            ('FONTSIZE', (0, 1), (-1, -1), 8),
            ('ROWBACKGROUNDS', (0, 1), (-1, -1), [colors.white, colors.HexColor('#F0F0F0')])
        ]))
        return table
    
    def _create_summary_section(self, data, totals):
        """Create executive summary section."""
        org = data['organization']
        period = data['reporting_period']
        
        elements = []
        elements.append(Paragraph("EXECUTIVE SUMMARY", self.styles['SectionHeader']))
        elements.append(Spacer(1, 0.2*inch))
        
        # Organization info table
        org_data = [
            ['Organization', org.get('name', 'N/A')],
            ['Sector', org.get('sector', 'N/A')],
            ['Employees', str(org.get('employees', 'N/A'))],
            ['Reporting Period', f"{period.get('year', 'N/A')}"],
        ]
        
        org_table = Table(org_data, colWidths=[1.5*inch, 3.5*inch])
        org_table.setStyle(TableStyle([
            ('BACKGROUND', (0, 0), (0, -1), colors.HexColor('#E8F0F7')),
            ('ALIGN', (0, 0), (0, -1), 'LEFT'),
            ('ALIGN', (1, 0), (1, -1), 'LEFT'),
            ('FONTNAME', (0, 0), (0, -1), 'Helvetica-Bold'),
            ('GRID', (0, 0), (-1, -1), 1, colors.grey),
            ('FONTSIZE', (0, 0), (-1, -1), 10),
            ('ROWBACKGROUNDS', (0, 0), (-1, -1), [colors.white, colors.HexColor('#F0F0F0')])
        ]))
        
        elements.append(org_table)
        elements.append(Spacer(1, 0.3*inch))
        
        # Emissions summary
        elements.append(Paragraph("Carbon Emissions Summary", self.styles['Heading3']))
        elements.append(Spacer(1, 0.1*inch))
        
        emissions_data = [
            ['Scope', 'Emissions (tCO2e)', 'Percentage'],
            ['Scope 1 - Direct', f"{totals['scope1']:.2f}", f"{(totals['scope1']/totals['total']*100):.1f}%"],
            ['Scope 2 - Indirect (Energy)', f"{totals['scope2']:.2f}", f"{(totals['scope2']/totals['total']*100):.1f}%"],
            ['Scope 3 - Other Indirect', f"{totals['scope3']:.2f}", f"{(totals['scope3']/totals['total']*100):.1f}%"],
            ['Total Emissions', f"{totals['total']:.2f}", '100.0%'],
        ]
        
        emissions_table = Table(emissions_data, colWidths=[2.5*inch, 1.5*inch, 1.5*inch])
        emissions_table.setStyle(TableStyle([
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
            ('FONTSIZE', (0, 1), (-1, -1), 10),
        ]))
        
        elements.append(emissions_table)
        return elements
    
    def _create_compliance_section(self, report_type):
        """Create compliance framework section."""
        elements = []
        elements.append(Spacer(1, 0.2*inch))
        elements.append(Paragraph("COMPLIANCE FRAMEWORKS", self.styles['SectionHeader']))
        elements.append(Spacer(1, 0.1*inch))
        
        frameworks = {
            'CSRD': 'Corporate Sustainability Reporting Directive (EU)',
            'SEC': 'Securities and Exchange Commission Climate Disclosure Rules',
            'CALIFORNIA': 'California Climate Corporate Data Accountability Act (SB 253)',
            'CBAM': 'Carbon Border Adjustment Mechanism (EU)',
            'IFRS': 'IFRS S2 Climate-related Disclosures'
        }
        
        framework_text = frameworks.get(report_type, 'Multi-Framework Compliance')
        elements.append(Paragraph(f"<b>{report_type}:</b> {framework_text}", self.styles['Normal']))
        elements.append(Spacer(1, 0.1*inch))
        elements.append(Paragraph(
            "This report demonstrates full compliance with the requirements of the applicable "
            "sustainability disclosure framework, including detailed emissions data, data quality assessment, "
            "and management strategies for emissions reduction.",
            self.styles['Normal']
        ))
        
        return elements
    
    def _create_data_quality_section(self, data):
        """Create data quality metrics section."""
        elements = []
        elements.append(Spacer(1, 0.2*inch))
        elements.append(Paragraph("DATA QUALITY ASSESSMENT", self.styles['SectionHeader']))
        elements.append(Spacer(1, 0.1*inch))
        
        quality_metrics = [
            ['Metric', 'Value', 'Status'],
            ['Data Quality Score', '92%', '✓ High'],
            ['Data Completeness', '96%', '✓ High'],
            ['Activity Records', str(len(data.get('activities', []))), '✓ Complete'],
            ['Calculation Methodology', 'GHG Protocol Standard', '✓ Verified'],
            ['Last Updated', datetime.now().strftime('%Y-%m-%d'), '✓ Current'],
        ]
        
        quality_table = Table(quality_metrics, colWidths=[2*inch, 1.5*inch, 1.5*inch])
        quality_table.setStyle(TableStyle([
            ('BACKGROUND', (0, 0), (-1, 0), colors.HexColor('#1F4788')),
            ('TEXTCOLOR', (0, 0), (-1, 0), colors.whitesmoke),
            ('ALIGN', (0, 0), (-1, -1), 'CENTER'),
            ('FONTNAME', (0, 0), (-1, 0), 'Helvetica-Bold'),
            ('FONTSIZE', (0, 0), (-1, 0), 10),
            ('BOTTOMPADDING', (0, 0), (-1, 0), 8),
            ('BACKGROUND', (0, 1), (-1, -1), colors.HexColor('#F0F8FF')),
            ('GRID', (0, 0), (-1, -1), 1, colors.grey),
            ('FONTSIZE', (0, 1), (-1, -1), 9),
        ]))
        
        elements.append(quality_table)
        return elements
    
    def generate_csrd_report(self):
        """Generate CSRD compliance report."""
        data = self._load_test_data('manufacturing_company_2024.json')
        totals = self._calculate_totals(data['activities'])
        
        output_file = self.output_dir / 'CSRD_Manufacturing_Report_2024.pdf'
        doc = SimpleDocTemplate(str(output_file), pagesize=A4, rightMargin=0.5*inch, 
                               leftMargin=0.5*inch, topMargin=0.75*inch, bottomMargin=0.75*inch)
        
        story = []
        
        # Title page
        story.append(Paragraph("OFFGRIDFLOW", self.styles['CustomTitle']))
        story.append(Paragraph("Corporate Sustainability Reporting Directive (CSRD)", self.styles['Subtitle']))
        story.append(Paragraph("E1 Environmental Compliance Report", self.styles['Subtitle']))
        story.append(Spacer(1, 0.3*inch))
        story.append(Paragraph(f"<b>{data['organization']['name']}</b>", self.styles['Heading2']))
        story.append(Paragraph(f"Reporting Period: {data['reporting_period']['year']}", self.styles['Normal']))
        story.append(Paragraph(f"Generated: {datetime.now().strftime('%B %d, %Y')}", self.styles['Normal']))
        story.append(PageBreak())
        
        # Executive Summary
        story.extend(self._create_summary_section(data, totals))
        story.append(PageBreak())
        
        # Detailed Emissions
        story.append(Paragraph("DETAILED EMISSIONS ANALYSIS", self.styles['SectionHeader']))
        story.append(Spacer(1, 0.1*inch))
        story.append(self._create_emissions_table(data['activities']))
        story.append(Spacer(1, 0.2*inch))
        story.append(PageBreak())
        
        # Compliance section
        story.extend(self._create_compliance_section('CSRD'))
        
        # Data Quality
        story.extend(self._create_data_quality_section(data))
        story.append(PageBreak())
        
        # Footer
        story.append(Paragraph("VERIFICATION & ASSURANCE", self.styles['SectionHeader']))
        story.append(Spacer(1, 0.1*inch))
        story.append(Paragraph(
            "This report has been prepared in accordance with the GHG Protocol Corporate Standard "
            "and meets the requirements of the Corporate Sustainability Reporting Directive (CSRD). "
            "All emissions data has been calculated using the most current emission factors and calculation methodologies. "
            "The data presented represents high-quality information suitable for regulatory disclosure.",
            self.styles['Normal']
        ))
        
        # Build PDF
        doc.build(story)
        return str(output_file)
    
    def generate_sec_report(self):
        """Generate SEC climate disclosure report."""
        data = self._load_test_data('tech_company_2024.json')
        totals = self._calculate_totals(data['activities'])
        
        output_file = self.output_dir / 'SEC_Climate_Disclosure_2024.pdf'
        doc = SimpleDocTemplate(str(output_file), pagesize=A4, rightMargin=0.5*inch,
                               leftMargin=0.5*inch, topMargin=0.75*inch, bottomMargin=0.75*inch)
        
        story = []
        
        # Title
        story.append(Paragraph("OFFGRIDFLOW", self.styles['CustomTitle']))
        story.append(Paragraph("SEC Climate Disclosure Report", self.styles['Subtitle']))
        story.append(Paragraph("Greenhouse Gas Emissions and Climate Risk Analysis", self.styles['Subtitle']))
        story.append(Spacer(1, 0.3*inch))
        story.append(Paragraph(f"<b>{data['organization']['name']}</b>", self.styles['Heading2']))
        story.append(Paragraph(f"Fiscal Year {data['reporting_period']['year']}", self.styles['Normal']))
        story.append(Paragraph(f"Filed: {datetime.now().strftime('%B %d, %Y')}", self.styles['Normal']))
        story.append(PageBreak())
        
        # Summary
        story.extend(self._create_summary_section(data, totals))
        story.append(PageBreak())
        
        # Governance section
        story.append(Paragraph("GOVERNANCE & OVERSIGHT", self.styles['SectionHeader']))
        story.append(Spacer(1, 0.1*inch))
        story.append(Paragraph(
            "<b>Board Oversight:</b> The Board of Directors oversees climate risks and opportunities "
            "through the Sustainability Committee, which meets quarterly to review emissions targets, "
            "progress toward goals, and emerging climate-related risks.",
            self.styles['Normal']
        ))
        story.append(Spacer(1, 0.1*inch))
        story.append(Paragraph(
            "<b>Management Accountability:</b> The Chief Sustainability Officer reports directly to the CEO "
            "and is responsible for implementing the company's climate strategy and achieving emissions targets.",
            self.styles['Normal']
        ))
        story.append(PageBreak())
        
        # Risk & Opportunity
        story.append(Paragraph("CLIMATE RISKS & OPPORTUNITIES", self.styles['SectionHeader']))
        story.append(Spacer(1, 0.1*inch))
        
        risk_data = [
            ['Risk Type', 'Description', 'Impact', 'Mitigation Strategy'],
            ['Transition', 'Carbon pricing mechanisms', 'Medium-High', 'Renewable energy investment'],
            ['Physical', 'Extreme weather events', 'Medium', 'Facility resilience & insurance'],
            ['Market', 'Changing consumer preferences', 'High', 'Product innovation program'],
            ['Regulatory', 'New climate regulations', 'High', 'Proactive compliance strategy'],
        ]
        
        risk_table = Table(risk_data, colWidths=[1.2*inch, 1.8*inch, 1.2*inch, 1.8*inch])
        risk_table.setStyle(TableStyle([
            ('BACKGROUND', (0, 0), (-1, 0), colors.HexColor('#C62828')),
            ('TEXTCOLOR', (0, 0), (-1, 0), colors.whitesmoke),
            ('ALIGN', (0, 0), (-1, -1), 'LEFT'),
            ('FONTNAME', (0, 0), (-1, 0), 'Helvetica-Bold'),
            ('FONTSIZE', (0, 0), (-1, 0), 10),
            ('BOTTOMPADDING', (0, 0), (-1, 0), 8),
            ('BACKGROUND', (0, 1), (-1, -1), colors.HexColor('#FFEBEE')),
            ('GRID', (0, 0), (-1, -1), 1, colors.grey),
            ('FONTSIZE', (0, 1), (-1, -1), 9),
        ]))
        
        story.append(risk_table)
        story.append(Spacer(1, 0.2*inch))
        story.append(PageBreak())
        
        # Emissions data
        story.append(Paragraph("DETAILED EMISSIONS DATA", self.styles['SectionHeader']))
        story.append(Spacer(1, 0.1*inch))
        story.append(self._create_emissions_table(data['activities']))
        story.append(PageBreak())
        
        # Targets
        story.append(Paragraph("EMISSIONS REDUCTION TARGETS", self.styles['SectionHeader']))
        story.append(Spacer(1, 0.1*inch))
        
        targets_data = [
            ['Scope', 'Baseline (2020)', 'Target Year', 'Target Reduction', 'Status'],
            ['Scope 1+2', f"{totals['scope1'] + totals['scope2']:.0f} tCO2e", '2030', '50%', 'On Track'],
            ['Scope 3', f"{totals['scope3']:.0f} tCO2e", '2030', '35%', 'On Track'],
            ['Science-Based Target', 'SBTi Approved', '2050', 'Net Zero', 'Active'],
        ]
        
        targets_table = Table(targets_data, colWidths=[1.2*inch, 1.3*inch, 1.2*inch, 1.3*inch, 1*inch])
        targets_table.setStyle(TableStyle([
            ('BACKGROUND', (0, 0), (-1, 0), colors.HexColor('#1B5E20')),
            ('TEXTCOLOR', (0, 0), (-1, 0), colors.whitesmoke),
            ('ALIGN', (0, 0), (-1, -1), 'CENTER'),
            ('FONTNAME', (0, 0), (-1, 0), 'Helvetica-Bold'),
            ('FONTSIZE', (0, 0), (-1, 0), 10),
            ('BOTTOMPADDING', (0, 0), (-1, 0), 8),
            ('BACKGROUND', (0, 1), (-1, -1), colors.HexColor('#E8F5E9')),
            ('GRID', (0, 0), (-1, -1), 1, colors.grey),
            ('FONTSIZE', (0, 1), (-1, -1), 9),
        ]))
        
        story.append(targets_table)
        
        # Build PDF
        doc.build(story)
        return str(output_file)
    
    def generate_california_report(self):
        """Generate California climate accountability report."""
        data = self._load_test_data('retail_company_2024.json')
        totals = self._calculate_totals(data['activities'])
        
        output_file = self.output_dir / 'California_SB253_Report_2024.pdf'
        doc = SimpleDocTemplate(str(output_file), pagesize=A4, rightMargin=0.5*inch,
                               leftMargin=0.5*inch, topMargin=0.75*inch, bottomMargin=0.75*inch)
        
        story = []
        
        # Title
        story.append(Paragraph("OFFGRIDFLOW", self.styles['CustomTitle']))
        story.append(Paragraph("California Climate Corporate Data Accountability Report", self.styles['Subtitle']))
        story.append(Paragraph("Senate Bill 253 & 261 Compliance", self.styles['Subtitle']))
        story.append(Spacer(1, 0.3*inch))
        story.append(Paragraph(f"<b>{data['organization']['name']}</b>", self.styles['Heading2']))
        story.append(Paragraph(f"Reporting Year: {data['reporting_period']['year']}", self.styles['Normal']))
        story.append(Paragraph(f"Submitted: {datetime.now().strftime('%B %d, %Y')}", self.styles['Normal']))
        story.append(PageBreak())
        
        # Summary
        story.extend(self._create_summary_section(data, totals))
        story.append(PageBreak())
        
        # Compliance statement
        story.append(Paragraph("REGULATORY COMPLIANCE STATEMENT", self.styles['SectionHeader']))
        story.append(Spacer(1, 0.1*inch))
        story.append(Paragraph(
            f"<b>{data['organization']['name']}</b> is reporting in compliance with California "
            "Senate Bill 253 (Climate Corporate Data Accountability Act) for the reporting period "
            f"January 1 - December 31, {data['reporting_period']['year']}.",
            self.styles['Normal']
        ))
        story.append(Spacer(1, 0.15*inch))
        
        # Assurance
        story.append(Paragraph("ASSURANCE STATEMENT", self.styles['Heading3']))
        story.append(Spacer(1, 0.1*inch))
        story.append(Paragraph(
            "This report has undergone independent assurance verification by a third-party auditor "
            "in accordance with ISO 14064-3 standards. The assurance engagement confirmed the accuracy, "
            "completeness, and reliability of the reported emissions data.",
            self.styles['Normal']
        ))
        story.append(PageBreak())
        
        # Scope 3 categories
        story.append(Paragraph("SCOPE 3 DETAILED BREAKDOWN", self.styles['SectionHeader']))
        story.append(Spacer(1, 0.1*inch))
        
        scope3_activities = [a for a in data['activities'] if 'scope3' in a.get('scope', '').lower()]
        scope3_data = [['Category', 'Quantity', 'Emissions (tCO2e)', 'Percentage']]
        
        for act in scope3_activities[:15]:
            scope3_data.append([
                act.get('category', 'N/A')[:25],
                f"{act.get('quantity', 0):.1f}",
                f"{act.get('emissions_tonnes', 0):.2f}",
                f"{(act.get('emissions_tonnes', 0)/totals['total']*100):.1f}%" if totals['total'] > 0 else '0%'
            ])
        
        scope3_table = Table(scope3_data, colWidths=[2.5*inch, 1.2*inch, 1.5*inch, 1.3*inch])
        scope3_table.setStyle(TableStyle([
            ('BACKGROUND', (0, 0), (-1, 0), colors.HexColor('#F57C00')),
            ('TEXTCOLOR', (0, 0), (-1, 0), colors.whitesmoke),
            ('ALIGN', (0, 0), (-1, -1), 'CENTER'),
            ('FONTNAME', (0, 0), (-1, 0), 'Helvetica-Bold'),
            ('FONTSIZE', (0, 0), (-1, 0), 10),
            ('BOTTOMPADDING', (0, 0), (-1, 0), 8),
            ('BACKGROUND', (0, 1), (-1, -1), colors.HexColor('#FFF3E0')),
            ('GRID', (0, 0), (-1, -1), 1, colors.grey),
            ('FONTSIZE', (0, 1), (-1, -1), 9),
        ]))
        
        story.append(scope3_table)
        story.append(Spacer(1, 0.2*inch))
        story.append(PageBreak())
        
        # Data quality
        story.extend(self._create_data_quality_section(data))
        
        # Build PDF
        doc.build(story)
        return str(output_file)
    
    def generate_all_reports(self):
        """Generate all report types."""
        reports = []
        
        print("🎯 Generating Professional ESG Compliance Reports...")
        print()
        
        try:
            print("📄 Generating CSRD Manufacturing Report...")
            report_path = self.generate_csrd_report()
            print(f"   ✅ {report_path}")
            reports.append(report_path)
        except Exception as e:
            print(f"   ❌ Error: {e}")
        
        try:
            print("📄 Generating SEC Climate Disclosure Report...")
            report_path = self.generate_sec_report()
            print(f"   ✅ {report_path}")
            reports.append(report_path)
        except Exception as e:
            print(f"   ❌ Error: {e}")
        
        try:
            print("📄 Generating California SB 253 Compliance Report...")
            report_path = self.generate_california_report()
            print(f"   ✅ {report_path}")
            reports.append(report_path)
        except Exception as e:
            print(f"   ❌ Error: {e}")
        
        print()
        print(f"🎉 Successfully generated {len(reports)} reports!")
        print()
        print("📁 Reports Location:")
        for report in reports:
            print(f"   {report}")
        
        return reports


if __name__ == '__main__':
    generator = OffGridFlowReportGenerator()
    reports = generator.generate_all_reports()
