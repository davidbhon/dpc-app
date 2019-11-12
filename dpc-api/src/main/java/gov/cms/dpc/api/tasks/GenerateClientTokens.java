package gov.cms.dpc.api.tasks;

import com.github.nitram509.jmacaroons.Macaroon;
import com.github.nitram509.jmacaroons.MacaroonVersion;
import com.google.common.collect.ImmutableCollection;
import com.google.common.collect.ImmutableMultimap;
import gov.cms.dpc.api.auth.OrganizationPrincipal;
import gov.cms.dpc.api.entities.TokenEntity;
import gov.cms.dpc.api.resources.v1.TokenResource;
import gov.cms.dpc.macaroons.MacaroonBakery;
import io.dropwizard.servlets.tasks.Task;
import org.hl7.fhir.dstu3.model.Organization;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import javax.inject.Inject;
import javax.inject.Singleton;
import java.io.PrintWriter;
import java.nio.charset.StandardCharsets;
import java.util.Base64;
import java.util.Collections;
import java.util.Optional;

/**
 * Admin task for creating a Golden macaroon which has superuser permissions in the application.
 * This should only ever be called once per environment.
 */
@Singleton
public class GenerateClientTokens extends Task {

    private static final Logger logger = LoggerFactory.getLogger(GenerateClientTokens.class);

    private final MacaroonBakery bakery;
    private final TokenResource resource;
    private final Base64.Encoder encoder;

    @Inject
    GenerateClientTokens(MacaroonBakery bakery, TokenResource resource) {
        super("generate-token");
        this.bakery = bakery;
        this.resource = resource;
        this.encoder = Base64.getUrlEncoder();
    }

    @Override
    public void execute(ImmutableMultimap<String, String> parameters, PrintWriter output) {

        final ImmutableCollection<String> organizationCollection = parameters.get("organization");
        if (organizationCollection.isEmpty()) {
            logger.warn("CREATING UNRESTRICTED MACAROON. ENSURE THIS IS OK");
            final Macaroon macaroon = bakery.createMacaroon(Collections.emptyList());
            final byte[] serialized = macaroon.serialize(MacaroonVersion.SerializationVersion.V1_BINARY).getBytes(StandardCharsets.UTF_8);
            output.write(this.encoder.encodeToString(serialized));
        } else {
            final String organization = organizationCollection.asList().get(0);
            final Organization orgResource = new Organization();
            orgResource.setId(organization);
            final TokenEntity tokenResponse = this.resource
                    .createOrganizationToken(
                            new OrganizationPrincipal(orgResource),
                            null,
                            Optional.empty());

            output.write(tokenResponse.getToken());
        }
    }
}
