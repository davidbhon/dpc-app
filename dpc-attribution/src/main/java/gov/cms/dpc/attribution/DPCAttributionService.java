package gov.cms.dpc.attribution;

import ca.mestevens.java.configuration.bundle.TypesafeConfigurationBundle;
import com.codahale.metrics.jersey2.InstrumentedResourceMethodApplicationListener;
import com.hubspot.dropwizard.guicier.GuiceBundle;
import com.squarespace.jersey2.guice.JerseyGuiceUtils;
import gov.cms.dpc.attribution.cli.SeedCommand;
import gov.cms.dpc.common.hibernate.DPCHibernateBundle;
import gov.cms.dpc.common.hibernate.DPCHibernateModule;
import gov.cms.dpc.common.utils.EnvironmentParser;
import gov.cms.dpc.fhir.FHIRModule;
import gov.cms.dpc.macaroons.BakeryModule;
import io.dropwizard.Application;
import io.dropwizard.db.PooledDataSourceFactory;
import io.dropwizard.migrations.MigrationsBundle;
import io.dropwizard.setup.Bootstrap;
import io.dropwizard.setup.Environment;
import io.federecio.dropwizard.swagger.SwaggerBundle;
import io.federecio.dropwizard.swagger.SwaggerBundleConfiguration;
import liquibase.exception.DatabaseException;
import org.knowm.dropwizard.sundial.SundialBundle;
import org.knowm.dropwizard.sundial.SundialConfiguration;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.sql.SQLException;
import java.util.List;

public class DPCAttributionService extends Application<DPCAttributionConfiguration> {

    private static final Logger logger = LoggerFactory.getLogger(DPCAttributionService.class);

    private final DPCHibernateBundle<DPCAttributionConfiguration> hibernateBundle = new DPCHibernateBundle<>(List.of("gov.cms.dpc.macaroons.store.hibernate.entities"));

    public static void main(final String[] args) throws Exception {
        new DPCAttributionService().run(args);
    }

    @Override
    public String getName() {
        return "DPC Attribution Service";
    }

    @Override
    public void initialize(Bootstrap<DPCAttributionConfiguration> bootstrap) {
        // This is required for Guice to load correctly. Not entirely sure why
        // https://github.com/dropwizard/dropwizard/issues/1772
        JerseyGuiceUtils.reset();
        registerBundles(bootstrap);

        bootstrap.addCommand(new SeedCommand(bootstrap.getApplication()));
    }

    @Override
    public void run(DPCAttributionConfiguration configuration, Environment environment) throws DatabaseException, SQLException {
        EnvironmentParser.getEnvironment("Attribution");
        final var listener = new InstrumentedResourceMethodApplicationListener(environment.metrics());
        environment.jersey().getResourceConfig().register(listener);
    }

    private void registerBundles(Bootstrap<DPCAttributionConfiguration> bootstrap) {
        GuiceBundle<DPCAttributionConfiguration> guiceBundle = GuiceBundle.defaultBuilder(DPCAttributionConfiguration.class)
                .modules(
                        new DPCHibernateModule<>(hibernateBundle),
                        new AttributionAppModule(hibernateBundle),
                        new FHIRModule<>(),
                        new BakeryModule())
                .build();

        // The Hibernate bundle must be initialized before Guice.
        // The Hibernate Guice module requires an initialized SessionFactory,
        // so Dropwizard needs to initialize the HibernateBundle first to create the SessionFactory.
        bootstrap.addBundle(hibernateBundle);

        bootstrap.addBundle(guiceBundle);
        bootstrap.addBundle(new TypesafeConfigurationBundle("dpc.attribution"));
        bootstrap.addBundle(new MigrationsBundle<DPCAttributionConfiguration>() {
            @Override
            public PooledDataSourceFactory getDataSourceFactory(DPCAttributionConfiguration configuration) {
                logger.debug("Connecting to database {} at {}", configuration.getDatabase().getDriverClass(), configuration.getDatabase().getUrl());
                return configuration.getDatabase();
            }
        });

        final SundialBundle<DPCAttributionConfiguration> sundialBundle = new SundialBundle<>() {
            @Override
            public SundialConfiguration getSundialConfiguration(DPCAttributionConfiguration dpcAttributionConfiguration) {
                return dpcAttributionConfiguration.getSundial();
            }
        };

        bootstrap.addBundle(sundialBundle);
        bootstrap.addBundle(new SwaggerBundle<DPCAttributionConfiguration>() {
            @Override
            protected SwaggerBundleConfiguration getSwaggerBundleConfiguration(DPCAttributionConfiguration configuration) {
                return configuration.getSwaggerBundleConfiguration();
            }
        });
    }
}
