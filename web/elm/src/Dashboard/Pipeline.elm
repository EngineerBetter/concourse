module Dashboard.Pipeline exposing
    ( hdPipelineView
    , pipelineCardStatus
    , pipelineNotSetView
    , pipelineStatus
    , pipelineView
    )

import Assets
import Concourse
import Concourse.BuildStatus exposing (BuildStatus(..))
import Concourse.PipelineStatus as PipelineStatus
import Dashboard.DashboardPreview as DashboardPreview
import Dashboard.Group.Models exposing (Pipeline, PipelineCardStatus(..))
import Dashboard.Styles as Styles
import Duration
import HoverState
import Html exposing (Html)
import Html.Attributes exposing (attribute, class, classList, draggable, href, style)
import Html.Events exposing (onClick, onMouseEnter, onMouseLeave)
import Message.Message exposing (DomID(..), Message(..))
import Routes
import Time
import UserState exposing (UserState)
import Views.Icon as Icon
import Views.PauseToggle as PauseToggle
import Views.Spinner as Spinner
import Views.Styles


pipelineNotSetView : Html Message
pipelineNotSetView =
    Html.div (class "card" :: Styles.noPipelineCard)
        [ Html.div
            (class "card-header" :: Styles.noPipelineCardHeader)
            [ Html.text "no pipeline set"
            ]
        , placeholderPreview
        , Html.div
            (class "card-footer" :: Styles.cardFooter)
            []
        ]


placeholderPreview : Html Message
placeholderPreview =
    Html.div
        (class "card-body" :: Styles.cardBody)
        [ Html.div Styles.previewPlaceholder [] ]


hdPipelineView :
    { pipeline : Pipeline
    , pipelineRunningKeyframes : String
    , resourceError : Bool
    , status : PipelineCardStatus
    }
    -> Html Message
hdPipelineView { pipeline, pipelineRunningKeyframes, resourceError, status } =
    Html.a
        ([ class "card"
         , attribute "data-pipeline-name" pipeline.name
         , attribute "data-team-name" pipeline.teamName
         , onMouseEnter <| TooltipHd pipeline.name pipeline.teamName
         , href <| Routes.toString <| Routes.pipelineRoute pipeline
         ]
            ++ Styles.pipelineCardHd status
        )
    <|
        [ Html.div
            (Styles.pipelineCardBannerHd
                { status = status
                , pipelineRunningKeyframes = pipelineRunningKeyframes
                }
            )
            []
        , Html.div
            (class "dashboardhd-pipeline-name" :: Styles.pipelineCardBodyHd)
            [ Html.text pipeline.name ]
        ]
            ++ (if resourceError then
                    [ Html.div Styles.resourceErrorTriangle [] ]

                else
                    []
               )


pipelineView :
    { now : Maybe Time.Posix
    , pipeline : Pipeline
    , hovered : HoverState.HoverState
    , pipelineRunningKeyframes : String
    , userState : UserState
    , resourceError : Bool
    , layers : List (List Concourse.Job)
    , query : String
    , isCached : Bool
    , status : PipelineCardStatus
    }
    -> Html Message
pipelineView { now, pipeline, hovered, pipelineRunningKeyframes, userState, resourceError, layers, query, isCached, status } =
    Html.div
        (Styles.pipelineCard
            ++ (if not isCached && String.isEmpty query then
                    [ style "cursor" "move" ]

                else
                    []
               )
            ++ (if isCached then
                    [ style "opacity" "0.45" ]

                else
                    []
               )
        )
        [ Html.div
            (class "banner"
                :: Styles.pipelineCardBanner
                    { status = status
                    , pipelineRunningKeyframes = pipelineRunningKeyframes
                    }
            )
            []
        , headerView pipeline resourceError
        , if status == PipelineStatusJobsDisabled then
            placeholderPreview

          else
            bodyView hovered layers
        , footerView userState pipeline now hovered status
        ]


pipelineCardStatus : Bool -> Bool -> List Concourse.Job -> Pipeline -> PipelineCardStatus
pipelineCardStatus isCached isJobsDisabled jobs pipeline =
    if isCached then
        PipelineStatusUnknown

    else if pipeline.paused then
        PipelineStatusPaused

    else if isJobsDisabled then
        PipelineStatusJobsDisabled

    else
        statusFromJobs jobs |> fromPipelineStatus


fromPipelineStatus : PipelineStatus.PipelineStatus -> PipelineCardStatus
fromPipelineStatus status =
    case status of
        PipelineStatus.PipelineStatusPaused ->
            PipelineStatusPaused

        PipelineStatus.PipelineStatusAborted details ->
            PipelineStatusAborted details

        PipelineStatus.PipelineStatusErrored details ->
            PipelineStatusErrored details

        PipelineStatus.PipelineStatusFailed details ->
            PipelineStatusFailed details

        PipelineStatus.PipelineStatusPending isRunning ->
            PipelineStatusPending isRunning

        PipelineStatus.PipelineStatusSucceeded details ->
            PipelineStatusSucceeded details


pipelineStatus : List Concourse.Job -> Pipeline -> PipelineStatus.PipelineStatus
pipelineStatus jobs pipeline =
    if pipeline.paused then
        PipelineStatus.PipelineStatusPaused

    else
        statusFromJobs jobs


statusFromJobs : List Concourse.Job -> PipelineStatus.PipelineStatus
statusFromJobs jobs =
    let
        isRunning =
            List.any (\job -> job.nextBuild /= Nothing) jobs

        mostImportantJobStatus =
            jobs
                |> List.map jobStatus
                |> List.sortWith Concourse.BuildStatus.ordering
                |> List.head

        firstNonSuccess =
            jobs
                |> List.filter (jobStatus >> (/=) BuildStatusSucceeded)
                |> List.filterMap transition
                |> List.sortBy Time.posixToMillis
                |> List.head

        lastTransition =
            jobs
                |> List.filterMap transition
                |> List.sortBy Time.posixToMillis
                |> List.reverse
                |> List.head

        transitionTime =
            case firstNonSuccess of
                Just t ->
                    Just t

                Nothing ->
                    lastTransition
    in
    case ( mostImportantJobStatus, transitionTime ) of
        ( _, Nothing ) ->
            PipelineStatus.PipelineStatusPending isRunning

        ( Nothing, _ ) ->
            PipelineStatus.PipelineStatusPending isRunning

        ( Just BuildStatusPending, _ ) ->
            PipelineStatus.PipelineStatusPending isRunning

        ( Just BuildStatusStarted, _ ) ->
            PipelineStatus.PipelineStatusPending isRunning

        ( Just BuildStatusSucceeded, Just since ) ->
            if isRunning then
                PipelineStatus.PipelineStatusSucceeded PipelineStatus.Running

            else
                PipelineStatus.PipelineStatusSucceeded (PipelineStatus.Since since)

        ( Just BuildStatusFailed, Just since ) ->
            if isRunning then
                PipelineStatus.PipelineStatusFailed PipelineStatus.Running

            else
                PipelineStatus.PipelineStatusFailed (PipelineStatus.Since since)

        ( Just BuildStatusErrored, Just since ) ->
            if isRunning then
                PipelineStatus.PipelineStatusErrored PipelineStatus.Running

            else
                PipelineStatus.PipelineStatusErrored (PipelineStatus.Since since)

        ( Just BuildStatusAborted, Just since ) ->
            if isRunning then
                PipelineStatus.PipelineStatusAborted PipelineStatus.Running

            else
                PipelineStatus.PipelineStatusAborted (PipelineStatus.Since since)


jobStatus : Concourse.Job -> BuildStatus
jobStatus job =
    case job.finishedBuild of
        Just build ->
            build.status

        Nothing ->
            BuildStatusPending


transition : Concourse.Job -> Maybe Time.Posix
transition =
    .transitionBuild >> Maybe.andThen (.duration >> .finishedAt)


headerView : Pipeline -> Bool -> Html Message
headerView pipeline resourceError =
    Html.a
        [ href <| Routes.toString <| Routes.pipelineRoute pipeline, draggable "false" ]
        [ Html.div
            ([ class "card-header"
             , onMouseEnter <| Tooltip pipeline.name pipeline.teamName
             ]
                ++ Styles.pipelineCardHeader
            )
            [ Html.div
                (class "dashboard-pipeline-name" :: Styles.pipelineName)
                [ Html.text pipeline.name ]
            , Html.div
                [ classList
                    [ ( "dashboard-resource-error", resourceError )
                    ]
                ]
                []
            ]
        ]


bodyView : HoverState.HoverState -> List (List Concourse.Job) -> Html Message
bodyView hovered layers =
    Html.div
        (class "card-body" :: Styles.pipelineCardBody)
        [ DashboardPreview.view hovered layers ]


footerView :
    UserState
    -> Pipeline
    -> Maybe Time.Posix
    -> HoverState.HoverState
    -> PipelineCardStatus
    -> Html Message
footerView userState pipeline now hovered status =
    let
        spacer =
            Html.div [ style "width" "13.5px" ] []

        pipelineId =
            { pipelineName = pipeline.name
            , teamName = pipeline.teamName
            }
    in
    Html.div
        (class "card-footer" :: Styles.pipelineCardFooter)
        [ Html.div
            [ style "display" "flex" ]
            [ Icon.icon
                { sizePx = 20, image = Assets.PipelineStatusIcon status }
                Styles.pipelineStatusIcon
            , transitionView now status
            ]
        , Html.div
            [ style "display" "flex" ]
          <|
            List.intersperse spacer
                [ PauseToggle.view
                    { isPaused =
                        status == PipelineStatusPaused
                    , pipeline = pipelineId
                    , isToggleHovered =
                        HoverState.isHovered (PipelineButton pipelineId) hovered
                    , isToggleLoading = pipeline.isToggleLoading
                    , tooltipPosition = Views.Styles.Above
                    , margin = "0"
                    , userState = userState
                    }
                , visibilityView
                    { public = pipeline.public
                    , pipelineId = pipelineId
                    , isClickable =
                        UserState.isAnonymous userState
                            || UserState.isMember
                                { teamName = pipeline.teamName
                                , userState = userState
                                }
                    , isHovered =
                        HoverState.isHovered (VisibilityButton pipelineId) hovered
                    , isVisibilityLoading = pipeline.isVisibilityLoading
                    }
                ]
        ]


visibilityView :
    { public : Bool
    , pipelineId : Concourse.PipelineIdentifier
    , isClickable : Bool
    , isHovered : Bool
    , isVisibilityLoading : Bool
    }
    -> Html Message
visibilityView { public, pipelineId, isClickable, isHovered, isVisibilityLoading } =
    if isVisibilityLoading then
        Spinner.hoverableSpinner
            { sizePx = 20
            , margin = "0"
            , hoverable = Just <| VisibilityButton pipelineId
            }

    else
        Html.div
            (Styles.visibilityToggle
                { public = public
                , isClickable = isClickable
                , isHovered = isHovered
                }
                ++ [ onMouseEnter <| Hover <| Just <| VisibilityButton pipelineId
                   , onMouseLeave <| Hover Nothing
                   ]
                ++ (if isClickable then
                        [ onClick <| Click <| VisibilityButton pipelineId ]

                    else
                        []
                   )
            )
            (if isClickable && isHovered then
                [ Html.div
                    Styles.visibilityTooltip
                    [ Html.text <|
                        if public then
                            "hide pipeline"

                        else
                            "expose pipeline"
                    ]
                ]

             else
                []
            )


sinceTransitionText : PipelineStatus.StatusDetails -> Time.Posix -> String
sinceTransitionText details now =
    case details of
        PipelineStatus.Running ->
            "running"

        PipelineStatus.Since time ->
            Duration.format <| Duration.between time now


transitionView : Maybe Time.Posix -> PipelineCardStatus -> Html Message
transitionView t status =
    case ( status, t ) of
        ( PipelineStatusPaused, _ ) ->
            Html.div
                (class "build-duration"
                    :: Styles.pipelineCardTransitionAge status
                )
                [ Html.text "paused" ]

        ( PipelineStatusUnknown, _ ) ->
            Html.div
                (class "build-duration"
                    :: Styles.pipelineCardTransitionAge status
                )
                [ Html.text "loading..." ]

        ( PipelineStatusPending False, _ ) ->
            Html.div
                (class "build-duration"
                    :: Styles.pipelineCardTransitionAge status
                )
                [ Html.text "pending" ]

        ( PipelineStatusPending True, _ ) ->
            Html.div
                (class "build-duration"
                    :: Styles.pipelineCardTransitionAge status
                )
                [ Html.text "running" ]

        ( PipelineStatusAborted details, Just now ) ->
            Html.div
                (class "build-duration"
                    :: Styles.pipelineCardTransitionAge status
                )
                [ Html.text <| sinceTransitionText details now ]

        ( PipelineStatusErrored details, Just now ) ->
            Html.div
                (class "build-duration"
                    :: Styles.pipelineCardTransitionAge status
                )
                [ Html.text <| sinceTransitionText details now ]

        ( PipelineStatusFailed details, Just now ) ->
            Html.div
                (class "build-duration"
                    :: Styles.pipelineCardTransitionAge status
                )
                [ Html.text <| sinceTransitionText details now ]

        ( PipelineStatusSucceeded details, Just now ) ->
            Html.div
                (class "build-duration"
                    :: Styles.pipelineCardTransitionAge status
                )
                [ Html.text <| sinceTransitionText details now ]

        _ ->
            Html.text ""
