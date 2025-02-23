package gov.cms.dpc.aggregation;

import ca.uhn.fhir.context.FhirContext;
import ca.uhn.fhir.rest.client.api.IGenericClient;
import ca.uhn.fhir.rest.client.api.ServerValidationModeEnum;
import com.codahale.metrics.MetricRegistry;
import com.google.inject.Binder;
import com.google.inject.Provides;
import com.google.inject.name.Named;
import com.hubspot.dropwizard.guicier.DropwizardAwareModule;
import com.typesafe.config.Config;
import gov.cms.dpc.aggregation.engine.AggregationEngine;
import gov.cms.dpc.aggregation.engine.JobBatchProcessor;
import gov.cms.dpc.aggregation.engine.JobBatchProcessorV2;
import gov.cms.dpc.aggregation.engine.OperationsConfig;
import gov.cms.dpc.aggregation.service.*;
import gov.cms.dpc.common.annotations.ExportPath;
import gov.cms.dpc.common.annotations.JobTimeout;
import gov.cms.dpc.common.hibernate.attribution.DPCManagedSessionFactory;
import gov.cms.dpc.fhir.hapi.ContextUtils;
import gov.cms.dpc.queue.models.JobQueueBatch;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import javax.inject.Singleton;

public class AggregationAppModule extends DropwizardAwareModule<DPCAggregationConfiguration> {

    private static final Logger logger = LoggerFactory.getLogger(AggregationAppModule.class);


    AggregationAppModule() {
        // Not used
    }

    @Override
    public void configure(Binder binder) {
        binder.bind(AggregationEngine.class);
        binder.bind(AggregationManager.class).asEagerSingleton();
        binder.bind(JobBatchProcessor.class);
        binder.bind(JobBatchProcessorV2.class);

        // Healthchecks
        // Additional health-checks can be added here
        // By default, Dropwizard adds a check for Hibernate and each additonal database (e.g. auth, queue, etc)
        // We also have JobQueueHealthy which ensures the queue is operation correctly
        // We have the BlueButton Client healthcheck as well
    }

    @Provides
    @Singleton
    public FhirContext provideSTU3Context() {
        final var fhirContext = FhirContext.forDstu3();

        // Setup the context with model scans (avoids doing this on the fetch threads and perhaps multithreaded bug)
        ContextUtils.prefetchResourceModels(fhirContext, JobQueueBatch.validResourceTypes);
        return fhirContext;
    }

    @Provides
    @Singleton
    @Named("fhirContextR4")
    public FhirContext provideR4Context() {
        final var fhirContext = FhirContext.forR4();

        // Setup the context with model scans (avoids doing this on the fetch threads and perhaps multithreaded bug)
        ContextUtils.prefetchResourceModels(fhirContext, JobQueueBatch.validResourceTypes);
        return fhirContext;
    }

    @Provides
    @Singleton
    MetricRegistry provideMetricRegistry() {
        return getEnvironment().metrics();
    }

    @Provides
    public Config provideConfig() {
        return getConfiguration().getConfig();
    }

    @Provides
    @ExportPath
    public String provideExportPath() {
        return getConfiguration().getExportPath();
    }

    @Provides
    OperationsConfig provideOperationsConfig() {
        final var config = getConfiguration();

        return new OperationsConfig(
                config.getResourcesPerFileCount(),
                config.getExportPath(),
                config.getRetryCount(),
                config.getPollingFrequency(),
                config.getLookBackMonths(),
                config.getLookBackDate(),
                config.getLookBackExemptOrgs()
        );
    }

    @Provides
    @JobTimeout
    public int provideJobTimeoutInSeconds() {
        return getConfiguration().getJobTimeoutInSeconds();
    }

    @Provides
    LookBackService provideLookBackService(DPCManagedSessionFactory sessionFactory, OperationsConfig operationsConfig) {
        //Configuring to skip look back when look back months is less than 0
        if (operationsConfig.getLookBackMonths() < 0) {
            return new EveryoneGetsDataLookBackServiceImpl();
        }
        return new LookBackServiceImpl(operationsConfig);
    }

    @Provides
    @Singleton
    @Named("consentClient")
    public IGenericClient provideConsentClient(FhirContext ctx) {
        logger.info("Connecting to consent server at {}.", getConfiguration().getConsentServiceUrl());
        ctx.getRestfulClientFactory().setServerValidationMode(ServerValidationModeEnum.NEVER);
        IGenericClient client = ctx.newRestfulGenericClient(getConfiguration().getConsentServiceUrl());
        return client;
    }

    @Provides
    ConsentService provideConsentService(@Named("consentClient") IGenericClient consentClient) {
        return new ConsentServiceImpl(consentClient);
    }
}
