package gov.cms.dpc.api.resources.v1;

import ca.uhn.fhir.parser.IParser;
import ca.uhn.fhir.rest.api.MethodOutcome;
import ca.uhn.fhir.rest.client.api.IGenericClient;
import ca.uhn.fhir.rest.gclient.IOperationUntypedWithInput;
import ca.uhn.fhir.rest.gclient.IReadExecutable;
import ca.uhn.fhir.rest.gclient.IUpdateExecutable;
import ca.uhn.fhir.rest.server.exceptions.AuthenticationException;
import ca.uhn.fhir.rest.server.exceptions.UnprocessableEntityException;
import gov.cms.dpc.api.APITestHelpers;
import gov.cms.dpc.api.AbstractSecureApplicationTest;
import gov.cms.dpc.api.TestOrganizationContext;
import gov.cms.dpc.bluebutton.client.MockBlueButtonClient;
import gov.cms.dpc.common.utils.NPIUtil;
import gov.cms.dpc.common.utils.SeedProcessor;
import gov.cms.dpc.fhir.DPCIdentifierSystem;
import gov.cms.dpc.fhir.FHIRExtractors;
import gov.cms.dpc.fhir.helpers.FHIRHelpers;
import gov.cms.dpc.testing.APIAuthHelpers;
import org.apache.commons.lang3.tuple.Pair;
import org.apache.http.HttpHeaders;
import org.eclipse.jetty.http.HttpStatus;
import org.hl7.fhir.dstu3.model.*;
import org.hl7.fhir.instance.model.api.IBaseOperationOutcome;
import org.junit.jupiter.api.MethodOrderer;
import org.junit.jupiter.api.Order;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestMethodOrder;

import javax.ws.rs.HttpMethod;
import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.URISyntaxException;
import java.net.URL;
import java.security.GeneralSecurityException;
import java.security.PrivateKey;
import java.sql.Date;
import java.util.List;
import java.util.UUID;

import static gov.cms.dpc.api.APITestHelpers.ORGANIZATION_ID;
import static gov.cms.dpc.api.APITestHelpers.ORGANIZATION_NPI;
import static org.junit.jupiter.api.Assertions.*;

@TestMethodOrder(MethodOrderer.OrderAnnotation.class)
class PatientResourceTest extends AbstractSecureApplicationTest {

    public static final String PROVENANCE_FMT = "{ \"resourceType\": \"Provenance\", \"recorded\": \"" + DateTimeType.now().getValueAsString() + "\"," +
            " \"reason\": [ { \"system\": \"http://hl7.org/fhir/v3/ActReason\", \"code\": \"TREAT\"  } ], \"agent\": [ { \"role\": " +
            "[ { \"coding\": [ { \"system\":" + "\"http://hl7.org/fhir/v3/RoleClass\", \"code\": \"AGNT\" } ] } ], \"whoReference\": " +
            "{ \"reference\": \"Organization/ORGANIZATION_ID\" }, \"onBehalfOfReference\": { \"reference\": " +
            "\"Practitioner/PRACTITIONER_ID\" } } ] }";

    final IParser parser = ctx.newJsonParser();
    final IGenericClient attrClient = APITestHelpers.buildAttributionClient(ctx);

    PatientResourceTest() {
        // Not used
    }

    @Test
    @Order(1)
    public void testCreatePatientReturnsAppropriateHeaders() {
        IGenericClient client = APIAuthHelpers.buildAuthenticatedClient(ctx, getBaseURL(), ORGANIZATION_TOKEN, PUBLIC_KEY_ID, PRIVATE_KEY);
        Patient patient = APITestHelpers.createPatientResource("4S41C00AA00", APITestHelpers.ORGANIZATION_ID);

        MethodOutcome methodOutcome = client.create()
                .resource(patient)
                .encodedJson()
                .execute();

        String location = methodOutcome.getResponseHeaders().get("location").get(0);
        String date = methodOutcome.getResponseHeaders().get("last-modified").get(0);
        assertNotNull(location);
        assertNotNull(date);

        Patient foundPatient = client.read()
                .resource(Patient.class)
                .withUrl(location)
                .encodedJson()
                .execute();

        assertEquals(patient.getIdentifierFirstRep().getValue(), foundPatient.getIdentifierFirstRep().getValue());

        client.delete()
                .resource(foundPatient)
                .encodedJson()
                .execute();
    }

    @Test
    @Order(2)
    void ensurePatientsExist() throws IOException, URISyntaxException, GeneralSecurityException {
        IGenericClient client = APIAuthHelpers.buildAuthenticatedClient(ctx, getBaseURL(), ORGANIZATION_TOKEN, PUBLIC_KEY_ID, PRIVATE_KEY);
        APITestHelpers.setupPatientTest(client, parser);

        final Bundle patients = fetchPatients(client);
        assertEquals(100, patients.getTotal(), "Should have correct number of patients");

        final Bundle specificSearch = fetchPatientBundleByMBI(client, "4S41C00AA00");

        assertEquals(1, specificSearch.getTotal(), "Should have a single patient");

        // Fetch the patient directly
        final Patient foundPatient = (Patient) specificSearch.getEntryFirstRep().getResource();

        final Patient queriedPatient = client
                .read()
                .resource(Patient.class)
                .withId(foundPatient.getIdElement())
                .encodedJson()
                .execute();

        assertTrue(foundPatient.equalsDeep(queriedPatient), "Search and GET should be identical");

        // Create a new org and make sure it has no providers
        final String m2 = FHIRHelpers.registerOrganization(attrClient, parser, OTHER_ORG_ID, "1112111111", getAdminURL());
        // Submit a new public key to use for JWT flow
        final String keyID = "new-key";
        final Pair<UUID, PrivateKey> uuidPrivateKeyPair = APIAuthHelpers.generateAndUploadKey(keyID, OTHER_ORG_ID, GOLDEN_MACAROON, getBaseURL());

        // Update the authenticated client to use the new organization
        client = APIAuthHelpers.buildAuthenticatedClient(ctx, getBaseURL(), m2, uuidPrivateKeyPair.getLeft(), uuidPrivateKeyPair.getRight());

        final Bundle otherPatients = fetchPatients(client);

        assertEquals(0, otherPatients.getTotal(), "Should not have any patients");

        // Try to look for one of the other patients
        final IReadExecutable<Patient> fetchRequest = client
                .read()
                .resource(Patient.class)
                .withId(foundPatient.getId())
                .encodedJson();

        assertThrows(AuthenticationException.class, fetchRequest::execute, "Should not be authorized");

        // Search, and find nothing
        final Bundle otherSpecificSearch = client
                .search()
                .forResource(Patient.class)
                .where(Patient.IDENTIFIER.exactly().identifier(foundPatient.getIdentifierFirstRep().getValue()))
                .returnBundle(Bundle.class)
                .encodedJson()
                .execute();

        assertEquals(0, otherSpecificSearch.getTotal(), "Should have a specific provider");
    }

    @Test
    @Order(3)
    void testPatientRemoval() throws IOException, URISyntaxException, GeneralSecurityException {
        final String macaroon = FHIRHelpers.registerOrganization(attrClient, parser, ORGANIZATION_ID, ORGANIZATION_NPI, getAdminURL());
        final String keyLabel = "patient-deletion-key";
        final Pair<UUID, PrivateKey> uuidPrivateKeyPair = APIAuthHelpers.generateAndUploadKey(keyLabel, ORGANIZATION_ID, GOLDEN_MACAROON, getBaseURL());
        final IGenericClient client = APIAuthHelpers.buildAuthenticatedClient(ctx, getBaseURL(), macaroon, uuidPrivateKeyPair.getLeft(), uuidPrivateKeyPair.getRight());

        final Bundle patients = fetchPatients(client);

        assertEquals(100, patients.getTotal(), "Should have correct number of patients");

        // Try to remove one

        final Patient patient = (Patient) patients.getEntry().get(patients.getTotal() - 2).getResource();

        client
                .delete()
                .resource(patient)
                .encodedJson()
                .execute();

        // Make sure it's done

        final IReadExecutable<Patient> fetchRequest = client
                .read()
                .resource(Patient.class)
                .withId(patient.getId())
                .encodedJson();

        // TODO: DPC-433, this really should be NotFound, but we can't disambiguate between the two cases
        assertThrows(AuthenticationException.class, fetchRequest::execute, "Should not have found the resource");

        // Search again
        final Bundle secondSearch = fetchPatients(client);

        assertEquals(99, secondSearch.getTotal(), "Should have correct number of patients");
    }

    @Test
    @Order(4)
    void testPatientUpdating() throws IOException, URISyntaxException, GeneralSecurityException {
        final String macaroon = FHIRHelpers.registerOrganization(attrClient, parser, ORGANIZATION_ID, ORGANIZATION_NPI, getAdminURL());
        final String keyLabel = "patient-update-key";
        final Pair<UUID, PrivateKey> uuidPrivateKeyPair = APIAuthHelpers.generateAndUploadKey(keyLabel, ORGANIZATION_ID, GOLDEN_MACAROON, getBaseURL());
        final IGenericClient client = APIAuthHelpers.buildAuthenticatedClient(ctx, getBaseURL(), macaroon, uuidPrivateKeyPair.getLeft(), uuidPrivateKeyPair.getRight());

        final Bundle patients = fetchPatients(client);

        assertEquals(99, patients.getTotal(), "Should have correct number of patients");

        // Try to update one
        final Patient patient = (Patient) patients.getEntry().get(patients.getTotal() - 2).getResource();
        patient.setBirthDate(Date.valueOf("2000-01-01"));
        patient.setGender(Enumerations.AdministrativeGender.MALE);

        final MethodOutcome outcome = client
                .update()
                .resource(patient)
                .withId(patient.getId())
                .encodedJson()
                .execute();

        assertTrue(((Patient) outcome.getResource()).equalsDeep(patient), "Should have been updated correctly");

        // Try to update with invalid MBI
        Identifier mbiIdentifier = patient.getIdentifier().stream()
                .filter(i -> DPCIdentifierSystem.MBI.getSystem().equals(i.getSystem())).findFirst().orElseThrow();
        mbiIdentifier.setValue("not-a-valid-MBI");

        IUpdateExecutable update = client
                .update()
                .resource(patient)
                .withId(patient.getId());

        assertThrows(UnprocessableEntityException.class, update::execute);
    }

    @Test
    @Order(5)
    void testCreateInvalidPatient() throws IOException, URISyntaxException {
        URL url = new URL(getBaseURL() + "/Patient");
        HttpURLConnection conn = (HttpURLConnection) url.openConnection();
        conn.setRequestMethod(HttpMethod.POST);
        conn.setRequestProperty(HttpHeaders.CONTENT_TYPE, "application/fhir+json");

        APIAuthHelpers.AuthResponse auth = APIAuthHelpers.jwtAuthFlow(getBaseURL(), ORGANIZATION_TOKEN, PUBLIC_KEY_ID, PRIVATE_KEY);
        conn.setRequestProperty(HttpHeaders.AUTHORIZATION, "Bearer " + auth.accessToken);

        conn.setDoOutput(true);
        String reqBody = "{\"test\": \"test\"}";
        conn.getOutputStream().write(reqBody.getBytes());

        assertEquals(HttpStatus.BAD_REQUEST_400, conn.getResponseCode());

        try (BufferedReader reader = new BufferedReader(new InputStreamReader(conn.getErrorStream()))) {
            StringBuilder respBuilder = new StringBuilder();
            String respLine;
            while ((respLine = reader.readLine()) != null) {
                respBuilder.append(respLine.trim());
            }
            String resp = respBuilder.toString();
            assertTrue(resp.contains("\"resourceType\":\"OperationOutcome\""));
            assertTrue(resp.contains("Invalid JSON content"));
        }

        conn.disconnect();
    }

    @Test
    @Order(6)
    void testPatientEverythingWithoutGroupFetchesData() throws IOException, URISyntaxException, GeneralSecurityException {
        IGenericClient client = generateClient(ORGANIZATION_ID, ORGANIZATION_NPI, "patient-everything-key");
        APITestHelpers.setupPractitionerTest(client, parser);

        String mbi = MockBlueButtonClient.TEST_PATIENT_MBIS.get(2);
        Patient patient = fetchPatient(client, mbi);
        Practitioner practitioner = fetchPractitionerByNPI(client, "1234329724");
        final String patientId = FHIRExtractors.getEntityUUID(patient.getId()).toString();

        // Patient without Group should still return data
        Bundle result = client
                .operation()
                .onInstance(new IdType("Patient", patientId))
                .named("$everything")
                .withNoParameters(Parameters.class)
                .returnResourceType(Bundle.class)
                .useHttpGet()
                .withAdditionalHeader("X-Provenance", generateProvenance(ORGANIZATION_ID, practitioner.getId()))
                .execute();

        assertEquals(64, result.getTotal(), "Should have 64 entries in Bundle");

        // Unattributed organization (unauthorized) without Group
        client = generateClient(OTHER_ORG_ID, "1112111111", "patient-everything-key-1");
        practitioner = createRandomPractitionerForOrg(client, OTHER_ORG_ID);

        String provenance = generateProvenance(OTHER_ORG_ID, practitioner.getId());

        IOperationUntypedWithInput<Bundle> everythingOp = client
                .operation()
                .onInstance(new IdType("Patient", patientId))
                .named("$everything")
                .withNoParameters(Parameters.class)
                .returnResourceType(Bundle.class)
                .useHttpGet()
                .withAdditionalHeader("X-Provenance", provenance);

        assertThrows(AuthenticationException.class, everythingOp::execute, "Org should not be be able to export another org's patient");
    }

    @Test
    @Order(7)
    void testPatientEverythingWithGroupFetchesData() throws IOException, URISyntaxException, GeneralSecurityException {
        IGenericClient client = generateClient(ORGANIZATION_ID, ORGANIZATION_NPI, "patient-everything-key-2");
        APITestHelpers.setupPractitionerTest(client, parser);

        String mbi = MockBlueButtonClient.TEST_PATIENT_MBIS.get(2);
        Patient patient = fetchPatient(client, mbi);
        Practitioner practitioner = fetchPractitionerByNPI(client, "1234329724");
        final String patientId = FHIRExtractors.getEntityUUID(patient.getId()).toString();

        // Patient in Group should also return data
        Group group = SeedProcessor.createBaseAttributionGroup(FHIRExtractors.getProviderNPI(practitioner), ORGANIZATION_ID);
        Reference patientRef = new Reference("Patient/" + patientId);
        group.addMember().setEntity(patientRef);

        client
                .create()
                .resource(group)
                .withAdditionalHeader("X-Provenance", generateProvenance(ORGANIZATION_ID, practitioner.getId()))
                .encodedJson()
                .execute();

        Bundle result = client
                .operation()
                .onInstance(new IdType("Patient", patientId))
                .named("$everything")
                .withNoParameters(Parameters.class)
                .returnResourceType(Bundle.class)
                .useHttpGet()
                .withAdditionalHeader("X-Provenance", generateProvenance(ORGANIZATION_ID, practitioner.getId()))
                .execute();

        assertEquals(64, result.getTotal(), "Should have 64 entries in Bundle");
        for (Bundle.BundleEntryComponent bec : result.getEntry()) {
            List<ResourceType> resourceTypes = List.of(ResourceType.Coverage, ResourceType.ExplanationOfBenefit, ResourceType.Patient);
            assertTrue(resourceTypes.contains(bec.getResource().getResourceType()), "Resource type should be Coverage, EOB, or Patient");
        }

        // Unattributed organization (unauthorized) with Group
        client = generateClient(OTHER_ORG_ID, "1112111111", "patient-everything-key-3");
        practitioner = createRandomPractitionerForOrg(client, OTHER_ORG_ID);

        group = SeedProcessor.createBaseAttributionGroup(FHIRExtractors.getProviderNPI(practitioner), OTHER_ORG_ID);
        patientRef = new Reference("Patient/" + patientId);
        group.addMember().setEntity(patientRef);
    }

    @Test
    public void tesGetPatientByUUID() throws GeneralSecurityException, IOException, URISyntaxException {
        final TestOrganizationContext orgAContext = registerAndSetupNewOrg();
        final TestOrganizationContext orgBContext = registerAndSetupNewOrg();
        final IGenericClient orgAClient = APIAuthHelpers.buildAuthenticatedClient(ctx, getBaseURL(), orgAContext.getClientToken(), UUID.fromString(orgAContext.getPublicKeyId()), orgAContext.getPrivateKey());
        final IGenericClient orgBClient = APIAuthHelpers.buildAuthenticatedClient(ctx, getBaseURL(), orgBContext.getClientToken(), UUID.fromString(orgBContext.getPublicKeyId()), orgBContext.getPrivateKey());

        //Setup org A with a patient
        final Patient orgAPatient = APITestHelpers.submitNewPatient(orgAClient, APITestHelpers.createPatientResource("4S41C00AA00", orgAContext.getOrgId()));

        //Setup org B with a patient
        final Patient orgBPatient = APITestHelpers.submitNewPatient(orgBClient, APITestHelpers.createPatientResource("4S41C00AA00", orgBContext.getOrgId()));

        assertNotNull(fetchPatientById(orgAClient,orgAPatient.getId()), "Org should be able to retrieve their own patient.");
        assertNotNull(fetchPatientById(orgBClient,orgBPatient.getId()), "Org should be able to retrieve their own patient.");
        assertThrows(AuthenticationException.class, () -> fetchPatientById(orgAClient, orgBPatient.getId()), "Expected auth error when retrieving another org's patient.");
    }

    @Test
    public void testDeletePatient() throws GeneralSecurityException, IOException, URISyntaxException {
        final TestOrganizationContext orgAContext = registerAndSetupNewOrg();
        final TestOrganizationContext orgBContext = registerAndSetupNewOrg();
        final IGenericClient orgAClient = APIAuthHelpers.buildAuthenticatedClient(ctx, getBaseURL(), orgAContext.getClientToken(), UUID.fromString(orgAContext.getPublicKeyId()), orgAContext.getPrivateKey());
        final IGenericClient orgBClient = APIAuthHelpers.buildAuthenticatedClient(ctx, getBaseURL(), orgBContext.getClientToken(), UUID.fromString(orgBContext.getPublicKeyId()), orgBContext.getPrivateKey());

        //Setup org B with a patient
        final Patient orgBPatient = APITestHelpers.submitNewPatient(orgBClient, APITestHelpers.createPatientResource("4S41C00AA00", orgBContext.getOrgId()));

        assertThrows(AuthenticationException.class, () -> deletePatientById(orgAClient, orgBPatient.getId()), "Expected auth error when deleting another org's patient.");
        assertNull(deletePatientById(orgBClient,orgBPatient.getId()), "Org should be able to delete their own patient.");
    }

    private IGenericClient generateClient(String orgID, String orgNPI, String keyLabel) throws IOException, URISyntaxException, GeneralSecurityException {
        final String macaroon = FHIRHelpers.registerOrganization(attrClient, parser, orgID, orgNPI, getAdminURL());
        final Pair<UUID, PrivateKey> uuidPrivateKeyPair = APIAuthHelpers.generateAndUploadKey(keyLabel, orgID, GOLDEN_MACAROON, getBaseURL());
        return APIAuthHelpers.buildAuthenticatedClient(ctx, getBaseURL(), macaroon, uuidPrivateKeyPair.getLeft(), uuidPrivateKeyPair.getRight(), false, true);
    }

    private String generateProvenance(String orgID, String practitionerID) {
        return PROVENANCE_FMT.replaceAll("ORGANIZATION_ID", orgID).replace("PRACTITIONER_ID", practitionerID);
    }

    private Bundle fetchPatients(IGenericClient client) {
        return client
                .search()
                .forResource(Patient.class)
                .encodedJson()
                .returnBundle(Bundle.class)
                .execute();
    }

    private Patient fetchPatient(IGenericClient client, String mbi) {
        return (Patient) fetchPatientBundleByMBI(client, mbi).getEntry().get(0).getResource();
    }

    private Patient fetchPatientById(IGenericClient client, String id) {
        return client.read()
                .resource(Patient.class)
                .withId(id)
                .encodedJson()
                .execute();
    }

    private IBaseOperationOutcome deletePatientById(IGenericClient client, String id) {
        return client.delete()
                .resourceById(new IdType(id))
                .encodedJson()
                .execute();
    }

    private Bundle fetchPatientBundleByMBI(IGenericClient client, String mbi) {
        return client
                .search()
                .forResource(Patient.class)
                .where(Patient.IDENTIFIER.exactly().systemAndCode(DPCIdentifierSystem.MBI.getSystem(), mbi))
                .returnBundle(Bundle.class)
                .encodedJson()
                .execute();
    }

    private Practitioner fetchPractitionerByNPI(IGenericClient client, String npi) {
        Bundle practSearch = client
                .search()
                .forResource(Practitioner.class)
                .where(Practitioner.IDENTIFIER.exactly().code(npi))
                .returnBundle(Bundle.class)
                .encodedJson()
                .execute();
        return (Practitioner) practSearch.getEntry().get(0).getResource();
    }

    private Practitioner createRandomPractitionerForOrg(IGenericClient client, String orgId) {
        Practitioner practitioner = APITestHelpers.createPractitionerResource(NPIUtil.generateNPI(), orgId);
        MethodOutcome methodOutcome = client.create()
                .resource(practitioner)
                .encodedJson()
                .execute();
        return (Practitioner) methodOutcome.getResource();
    }

}
