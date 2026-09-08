package ingest

var (
	triageOp = reportSubmissionOp[TriageReportInput]{
		kind:         KindTriage,
		reportSchema: SchemaTriage,
		layerID:      func(in TriageReportInput) string { return in.LayerID },
		sourceRef:    func(in TriageReportInput) string { return in.SourceRef },
		recordedAt:   func(in TriageReportInput) string { return in.RecordedAt },
		convert: func(in TriageReportInput) (ConvertResult, error) {
			return ConvertTriageReport(in.Report, in.SourceRef, in.RecordedAt, in.FingerprintIndex), nil
		},
	}
	validationOp = reportSubmissionOp[ValidationReportInput]{
		kind:         KindValidation,
		reportSchema: SchemaValidation,
		layerID:      func(in ValidationReportInput) string { return in.LayerID },
		sourceRef:    func(in ValidationReportInput) string { return in.SourceRef },
		recordedAt:   func(in ValidationReportInput) string { return in.RecordedAt },
		convert: func(in ValidationReportInput) (ConvertResult, error) {
			return ConvertValidationReport(in.Report, in.SourceRef, in.RecordedAt, in.FingerprintIndex), nil
		},
	}
	verificationOp = reportSubmissionOp[VerificationReportInput]{
		kind:         KindVerification,
		reportSchema: SchemaVerification,
		layerID:      func(in VerificationReportInput) string { return in.LayerID },
		sourceRef:    func(in VerificationReportInput) string { return in.SourceRef },
		recordedAt:   func(in VerificationReportInput) string { return in.RecordedAt },
		convert: func(in VerificationReportInput) (ConvertResult, error) {
			return ConvertVerificationReport(in.Report, in.SourceRef, in.RecordedAt, in.FingerprintIndex), nil
		},
	}
	countersignOp = submissionOp[CountersignInput]{kind: KindCountersign}
	severityOp    = submissionOp[SeverityInput]{kind: KindSeverity}
)
