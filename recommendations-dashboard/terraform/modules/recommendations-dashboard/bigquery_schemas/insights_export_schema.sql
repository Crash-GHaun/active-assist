[
  {"description":"Represents what cloud entity type the recommendation was generated for - eg: project number, billing account\\n","mode":"NULLABLE","name":"cloud_entity_type","type":"STRING"},
  {"description":"Value of the project number or billing account id\\n","mode":"NULLABLE","name":"cloud_entity_id","type":"STRING"},
  {"description":"Name of recommendation. A project recommendation is represented as\\nprojects/[PROJECT_NUMBER]/locations/[LOCATION]/recommenders/[RECOMMENDER_ID]/recommendations/[RECOMMENDATION_ID]\\n","mode":"NULLABLE","name":"name","type":"STRING"},
  {"description":"Location for which this recommendation is generated\\n","mode":"NULLABLE","name":"location","type":"STRING"},
  {"description":"Recommender ID of the recommender that has produced this recommendation\\n","mode":"NULLABLE","name":"insight_type","type":"STRING"},
  {"description":"Contains an identifier for a subtype of recommendations produced for the\\nsame recommender. Subtype is a function of content and impact, meaning a\\nnew subtype will be added when either content or primary impact category\\nchanges.\\nExamples:\\nFor recommender = \"google.iam.policy.Recommender\",\\nrecommender_subtype can be one of \"REMOVE_ROLE\"/\"REPLACE_ROLE\"\\n","mode":"NULLABLE","name":"insight_subtype","type":"STRING"},
  {"description":"Contains the fully qualified resource names for resources changed by the\\noperations in this recommendation. This field is always populated. ex:\\n[//cloudresourcemanager.googleapis.com/projects/foo].\\n","mode":"REPEATED","name":"target_resources","type":"STRING"},
  {"description":"Required. Free-form human readable summary in English.\\nThe maximum length is 500 characters.\\n","mode":"NULLABLE","name":"description","type":"STRING"},
  {"description":"Output only. Last time this recommendation was refreshed by the system that created it in the first place.\\n","mode":"NULLABLE","name":"last_refresh_time","type":"TIMESTAMP"},
  {"description":"Category being targeted by the insight. Can be one of:\\nUnspecified category.\\nCATEGORY_UNSPECIFIED = Unspecified category.\\nCOST = The insight is related to cost.\\nSECURITY = The insight is related to security.\\nPERFORMANCE = The insight is related to performance.\\nMANAGEABILITY = The insight is related to manageability.;\\n","mode":"NULLABLE","name":"category","type":"STRING"},
  {"description":"Output only. The state of the recommendation:\\n  STATE_UNSPECIFIED:\\n    Default state. Don't use directly.\\n  ACTIVE:\\n    Recommendation is active and can be applied. Recommendations content can\\n    be updated by Google.\\n    ACTIVE recommendations can be marked as CLAIMED, SUCCEEDED, or FAILED.\\n  CLAIMED:\\n    Recommendation is in claimed state. Recommendations content is\\n    immutable and cannot be updated by Google.\\n    CLAIMED recommendations can be marked as CLAIMED, SUCCEEDED, or FAILED.\\n  SUCCEEDED:\\n    Recommendation is in succeeded state. Recommendations content is\\n    immutable and cannot be updated by Google.\\n    SUCCEEDED recommendations can be marked as SUCCEEDED, or FAILED.\\n  FAILED:\\n    Recommendation is in failed state. Recommendations content is immutable\\n    and cannot be updated by Google.\\n    FAILED recommendations can be marked as SUCCEEDED, or FAILED.\\n  DISMISSED:\\n    Recommendation is in dismissed state.\\n    DISMISSED recommendations can be marked as ACTIVE.\\n","mode":"NULLABLE","name":"state","type":"STRING"},
  {"description":"Ancestry for the recommendation entity\\n","fields":[
    {"description":"Organization to which the recommendation project\\n","mode":"NULLABLE","name":"organization_id","type":"STRING"},
    {"description":"Up to 5 levels of parent folders for the recommendation project\\n","mode":"REPEATED","name":"folder_ids","type":"STRING"}
  ],"mode":"NULLABLE","name":"ancestors","type":"RECORD"},
  {"description":"Insights associated with this recommendation. A project insight is represented as\\nprojects/[PROJECT_NUMBER]/locations/[LOCATION]/insightTypes/[INSIGHT_TYPE_ID]/insights/[insight_id]\\n","mode":"REPEATED","name":"associated_recommendations","type":"STRING"},
  {"description":"Additional details about the insight in JSON format\\nschema:\\n  fields:\\n  - name: content\\n    type: STRING\\n    description: |\\n      A struct of custom fields to explain the insight.\\n      Example: \"grantedPermissionsCount\": \"1000\"\\n  - name: observation_period\\n    type: TIMESTAMP\\n    description: |\\n      Observation period that led to the insight. The source data used to\\n      generate the insight ends at last_refresh_time and begins at\\n      (last_refresh_time - observation_period).\\n- name: state_metadata\\n  type: STRING\\n  description: |\\n    A map of metadata for the state, provided by user or automations systems.\\n","mode":"NULLABLE","name":"insight_details","type":"STRING"},
  {"description":"Severity of the insight:\\n  SEVERITY_UNSPECIFIED:\\n    Default unspecified severity. Don't use directly.\\n  LOW:\\n    Lowest severity.\\n  MEDIUM:\\n    Second lowest severity.\\n  HIGH:\\n    Second highest severity.\\n  CRITICAL:\\n    Highest severity.\\n","mode":"NULLABLE","name":"severity","type":"STRING"}
]
