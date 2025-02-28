import React from "react";
import MultiColumn from "@theme-original/Footer/Links/MultiColumn";
import Feedback from "../../../../components/Feedback";

export default function MultiColumnWrapper(props) {
  return (
    <>
      <MultiColumn {...props} />
      <Feedback />
    </>
  );
}
