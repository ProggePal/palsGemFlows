This scoping spec outlines the MVP for the Kemira Value Calculator, focusing on concrete implementation details and identifying open questions.

---

## Kemira Value Calculator Scoping Spec

**1) One-liner summary**
A web-based, branded value calculator enabling Kemira sales staff to interactively demonstrate chemical/energy savings and ROI to municipal customers.

**2) Goals (business)**
*   Enhance sales presentation professionalism and credibility.
*   Improve customer engagement via real-time scenario adjustments.
*   Streamline initial sales qualification with quick ROI estimations.
*   Standardize value communication across sales force.
*   *Success Metrics:* Sales staff adoption rate, qualitative sales feedback, customer engagement improvement, branding consistency.

**3) Users & roles**
*   **Kemira Sales Staff (Primary User, ~20 initial users):** Accesses, inputs parameters, presents results.
*   **Municipal Customers (Audience):** Views results, provides input for parameter adjustments (not direct UI users).

**4) In-scope / Out-of-scope**

**In-scope (MVP Value Calculator):**
*   Desktop web-based application.
*   Input fields for 4-5 key parameters (chemical usage, process variables).
*   Calculation logic for savings (€) and ROI estimations (ballpark).
*   Display of calculated savings and ROI.
*   Real-time calculation updates as inputs change.
*   Professional UI/UX adhering to Kemira branding.
*   Basic error handling for invalid/missing inputs.
*   User access control for 20 sales staff.

**Out-of-scope (MVP):**
*   Saving, printing, or exporting calculation results (e.g., PDF quote).
*   User authentication beyond basic access control (e.g., SSO).
*   Complex customer data uploads (e.g., CSV imports for customer systems).
*   Integration with SAP or other external systems.
*   Advanced calculators for paper machine processes.
*   Any components of a training portal (courses, videos, user tracking).
*   Extensive admin portal (e.g., usage statistics, quotation management).

**5) User journeys**
1.  **Kemira Sales Staff demonstrates value calculator to customer:**
    *   Sales staff logs into web app.
    *   Enters 4-5 key customer-specific input parameters.
    *   Views calculated savings (€) and ROI estimates update in real-time.
    *   Adjusts input parameters interactively with customer to show scenarios.
    *   Presents professional results screen to customer.

**6) Screens & UI notes**
*   **Value Calculator Screen:**
    *   Purpose: Input parameters and display results side-by-side or dynamically.
    *   Components: Numeric input fields (4-5 parameters), display elements for savings (€) and ROI. Buttons: "Calculate" (if not real-time), "Reset".
    *   UI Notes: Visually professional, Kemira branded (logo, colors, fonts), intuitive for live modification.
    *   Error States: Clear messages for invalid/missing inputs, illogical results (e.g., zero savings/negative ROI).

**7) Data model (entities + key fields)**
*   **Value Calculation Inputs:**
    *   `param_1_name`: String (e.g., `ChemicalUsageRate`)
    *   `param_1_value`: Numeric
    *   `param_2_name`: String (e.g., `ProcessEfficiencyFactor`)
    *   `param_2_value`: Numeric
    *   ... (up to 4-5 key parameters)
*   **Value Calculation Outputs:**
    *   `calculated_savings_euro`: Numeric
    *   `calculated_roi_percentage`: Numeric
*   **User (for access control):**
    *   `username`: String
    *   `password_hash`: String

**8) Integrations**
*   **MVP:** None (standalone tool).
*   **Source of Logic:** Existing Excel sheets (logic to be translated).
*   **Future (potential, out-of-scope for MVP):** SAP ERP for consuming customer data.

**9) Non-functional requirements**
*   **Performance:** Fast real-time calculations (<1 sec response) during customer interaction.
*   **Security:** Access control for 20 sales staff. Data security for calculations. GDPR compliant (platform level).
*   **UI/UX:** Professional, marketing-friendly, intuitive. Adheres to Kemira branding guidelines (logo, colors, fonts).
*   **Platform:** Desktop web-based. Responsive UI.
*   **Scalability:** Supports initial 20 users, future growth based on active monthly user model.

**10) Risks & assumptions**
*   **Risks:**
    *   Complexity of Excel logic is higher than anticipated, leading to scope creep.
    *   Delays in obtaining detailed branding guidelines or Excel sheets.
    *   Limited 2025 budget or approval timeline delays project start.
*   **Assumptions:**
    *   App Farm platform can adequately handle UI, data model, and calculation logic.
    *   Existing Excel calculations are translatable.
    *   "Ballpark" ROI estimates are sufficient for MVP.
    *   Minimal to no administrative functionality required for MVP.

**11) Open questions**
*   What are the precise "4-5 key input parameters" (names, data types, value ranges)? (Requires review of Excel sheets).
*   What are the exact calculation formulas/business rules for savings and ROI? (Requires anonymized Excel sheets under NDA).
*   Are there specific interdependencies or complex validation rules between input parameters?
*   What are the detailed Kemira branding guidelines (color palette, fonts, specific logo usage) and required UI assets?
*   What is the desired level of user management/authentication for the 20 sales staff (e.g., simple login, specific credentials per user, shared access)?
*   What is the maximum realistic budget Kemira is targeting for a Q1 2025 project start?
*   What defines a "minimal scope proof-of-concept" that could potentially start before Christmas (e.g., 1-2 parameters, static results)?
*   Beyond aesthetics, what are the primary functional limitations or frustrations with the current Excel-based calculators from the sales team's perspective?

**12) MVP milestones (3-6 steps)**
1.  **Discovery & Design (Week 1-2):** Review Excel logic (under NDA), define precise inputs/formulas, gather branding guidelines. Finalize wireframes for main screen.
2.  **Core Calculation Engine (Week 3-4):** Implement backend logic for all calculations and data validation.
3.  **UI Development & Integration (Week 5-6):** Develop primary input/results screen. Integrate UI with calculation engine. Implement basic Kemira branding elements.
4.  **Access & Initial Testing (Week 7):** Implement user access control for 20 sales staff. Conduct internal UAT with key sales reps, gather feedback.
5.  **Branding Polish & Deployment (Week 8):** Refine UI/UX to full Kemira branding standards. Prepare for and deploy to production environment.